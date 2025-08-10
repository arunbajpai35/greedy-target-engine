from django.shortcuts import render
from rest_framework import viewsets, status
from rest_framework.response import Response
from .models import File
from .serializers import FileSerializer

import hashlib
from django.db.models import Q, Sum
from rest_framework.decorators import action


# Create your views here.


class FileViewSet(viewsets.ModelViewSet):
    queryset = File.objects.all()
    serializer_class = FileSerializer

    def create(self, request, *args, **kwargs):
        file_obj = request.FILES.get('file')
        if not file_obj:
            return Response({'error': 'No file provided'}, status=status.HTTP_400_BAD_REQUEST)

        # Compute SHA-256 content hash efficiently
        sha256 = hashlib.sha256()
        for chunk in file_obj.chunks():
            sha256.update(chunk)
        content_hash = sha256.hexdigest()

        # Try to find an existing non-duplicate file with same content_hash
        existing = File.objects.filter(content_hash=content_hash, is_duplicate=False).first()

        if existing:
            # Create a logical reference record without storing the file content again
            data = {
                'file': existing.file.name,  # reuse file path
                'original_filename': file_obj.name,
                'file_type': getattr(file_obj, 'content_type', ''),
                'size': file_obj.size,
                'content_hash': content_hash,
                'is_duplicate': True,
                'duplicate_of': str(existing.id),
            }
            serializer = self.get_serializer(data=data)
            serializer.is_valid(raise_exception=True)
            # Manually assign the existing FileField file since we pass name above
            instance = File(
                file=existing.file.name,
                original_filename=data['original_filename'],
                file_type=data['file_type'],
                size=data['size'],
                content_hash=content_hash,
                is_duplicate=True,
                duplicate_of=existing,
            )
            instance.save()
            output = self.get_serializer(instance)
            headers = self.get_success_headers(output.data)
            return Response(output.data, status=status.HTTP_201_CREATED, headers=headers)

        # Reset file pointer after hashing before saving
        try:
            file_obj.seek(0)
        except Exception:
            pass

        # Not a duplicate; store file and metadata
        data = {
            'file': file_obj,
            'original_filename': file_obj.name,
            'file_type': getattr(file_obj, 'content_type', ''),
            'size': file_obj.size,
            'content_hash': content_hash,
            'is_duplicate': False,
            'duplicate_of': None,
        }

        serializer = self.get_serializer(data=data)
        serializer.is_valid(raise_exception=True)
        self.perform_create(serializer)

        headers = self.get_success_headers(serializer.data)
        return Response(serializer.data, status=status.HTTP_201_CREATED, headers=headers)

    def get_queryset(self):
        qs = super().get_queryset()
        params = self.request.query_params

        # Search by filename
        search = params.get('search')
        if search:
            qs = qs.filter(original_filename__icontains=search)

        # Filter by file type
        file_type = params.get('file_type')
        if file_type:
            qs = qs.filter(file_type__icontains=file_type)

        # Size range filtering (bytes)
        size_min = params.get('size_min')
        size_max = params.get('size_max')
        if size_min:
            try:
                qs = qs.filter(size__gte=int(size_min))
            except ValueError:
                pass
        if size_max:
            try:
                qs = qs.filter(size__lte=int(size_max))
            except ValueError:
                pass

        # Upload date filtering: expect ISO dates (YYYY-MM-DD) or datetime
        date_from = params.get('date_from')
        date_to = params.get('date_to')
        if date_from:
            qs = qs.filter(uploaded_at__date__gte=date_from)
        if date_to:
            qs = qs.filter(uploaded_at__date__lte=date_to)

        return qs

    @action(detail=False, methods=['get'])
    def stats(self, request):
        """Return storage savings stats."""
        total_files = File.objects.count()
        duplicates = File.objects.filter(is_duplicate=True).count()
        unique_files = File.objects.filter(is_duplicate=False).count()
        # Physical storage used equals sum of sizes of unique files only
        storage_physical = File.objects.filter(is_duplicate=False).aggregate(total=Sum('size'))['total'] or 0
        storage_logical = File.objects.aggregate(total=Sum('size'))['total'] or 0
        savings_bytes = storage_logical - storage_physical
        return Response({
            'total_files': total_files,
            'unique_files': unique_files,
            'duplicates': duplicates,
            'storage_logical_bytes': storage_logical,
            'storage_physical_bytes': storage_physical,
            'savings_bytes': savings_bytes,
        })
