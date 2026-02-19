import React, { useRef, useState, DragEvent } from 'react';
import './FileUpload.css';

interface FileUploadProps {
  onFileSelect: (file: File) => void;
  currentFile?: File | null;
  maxSizeMB?: number;
}

const FileUpload: React.FC<FileUploadProps> = ({
  onFileSelect,
  currentFile,
  maxSizeMB = 5,
}) => {
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [isDragging, setIsDragging] = useState(false);
  const [preview, setPreview] = useState<string | null>(null);

  const handleFile = (file: File) => {
    // Validate file type
    if (!file.type.startsWith('image/')) {
      alert('Please select an image file');
      return;
    }

    // Validate file size
    const maxSizeBytes = maxSizeMB * 1024 * 1024;
    if (file.size > maxSizeBytes) {
      alert(`File size must be less than ${maxSizeMB}MB`);
      return;
    }

    onFileSelect(file);

    // Create preview
    const reader = new FileReader();
    reader.onloadend = () => {
      setPreview(reader.result as string);
    };
    reader.readAsDataURL(file);
  };

  const handleClick = () => {
    fileInputRef.current?.click();
  };

  const handleFileInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      handleFile(file);
    }
  };

  const handleDragOver = (e: DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    setIsDragging(true);
  };

  const handleDragLeave = (e: DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    setIsDragging(false);
  };

  const handleDrop = (e: DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    setIsDragging(false);

    const file = e.dataTransfer.files?.[0];
    if (file) {
      handleFile(file);
    }
  };

  return (
    <div className="file-upload">
      <div
        className={`file-upload-area ${isDragging ? 'dragging' : ''}`}
        onDragOver={handleDragOver}
        onDragLeave={handleDragLeave}
        onDrop={handleDrop}
        onClick={handleClick}
      >
        <input
          ref={fileInputRef}
          type="file"
          accept="image/*"
          onChange={handleFileInputChange}
          style={{ display: 'none' }}
        />
        {preview || currentFile ? (
          <div className="file-preview">
            <img src={preview || (currentFile ? URL.createObjectURL(currentFile) : '')} alt="Preview" />
            <button
              type="button"
              className="remove-file"
              onClick={(e) => {
                e.stopPropagation();
                setPreview(null);
                onFileSelect(null as any);
                if (fileInputRef.current) {
                  fileInputRef.current.value = '';
                }
              }}
            >
              ×
            </button>
          </div>
        ) : (
          <div className="file-upload-placeholder">
            <svg
              width="48"
              height="48"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
            >
              <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
              <polyline points="17 8 12 3 7 8" />
              <line x1="12" y1="3" x2="12" y2="15" />
            </svg>
            <p>Click or drag to upload</p>
            <p className="file-upload-hint">Max {maxSizeMB}MB</p>
          </div>
        )}
      </div>
    </div>
  );
};

export default FileUpload;
