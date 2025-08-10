from rest_framework import serializers
from .models import File


class FileSerializer(serializers.ModelSerializer):
    class Meta:
        model = File
        fields = [
            'id',
            'file',
            'original_filename',
            'file_type',
            'size',
            'uploaded_at',
            'content_hash',
            'is_duplicate',
            'duplicate_of',
        ]
        read_only_fields = ['id', 'uploaded_at'] 