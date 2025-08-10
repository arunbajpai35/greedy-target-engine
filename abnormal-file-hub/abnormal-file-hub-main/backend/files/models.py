from django.db import models
import uuid
import os


def file_upload_path(instance, filename):
    """Generate file path for new file upload"""
    ext = filename.split('.')[-1]
    filename = f"{uuid.uuid4()}.{ext}"
    return os.path.join('uploads', filename)


class File(models.Model):
    id = models.UUIDField(primary_key=True, default=uuid.uuid4, editable=False)
    file = models.FileField(upload_to=file_upload_path)
    original_filename = models.CharField(max_length=255, db_index=True)
    file_type = models.CharField(max_length=100, db_index=True)
    size = models.BigIntegerField(db_index=True)
    uploaded_at = models.DateTimeField(auto_now_add=True, db_index=True)

    # Deduplication fields
    content_hash = models.CharField(max_length=64, null=True, blank=True, db_index=True)
    is_duplicate = models.BooleanField(default=False, db_index=True)
    duplicate_of = models.ForeignKey(
        'self', null=True, blank=True, on_delete=models.SET_NULL, related_name='duplicates'
    )

    class Meta:
        ordering = ['-uploaded_at']

    def __str__(self):
        return self.original_filename
