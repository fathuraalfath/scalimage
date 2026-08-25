import React, { useState, useRef } from 'react';

const BACKEND_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

export default function CompressTool() {
  const [selectedFile, setSelectedFile] = useState(null);
  const [previewUrl, setPreviewUrl] = useState(null);
  const [targetFormat, setTargetFormat] = useState('webp');
  const [quality, setQuality] = useState(80);
  const [isProcessing, setIsProcessing] = useState(false);
  const fileInputRef = useRef(null);

  const handleFileSelect = (e) => {
    const file = e.target.files?.[0];
    if (file) {
      setSelectedFile(file);
      setPreviewUrl(URL.createObjectURL(file));
    }
  };

  const handleCompress = async () => {
    if (!selectedFile) return;

    setIsProcessing(true);

    // 1. Upload file first
    const formData = new FormData();
    formData.append('images', selectedFile);

    try {
      const uploadRes = await fetch(`${BACKEND_URL}/api/upload`, {
        method: 'POST',
        body: formData,
      });

      if (!uploadRes.ok) throw new Error(await uploadRes.text());
      const data = await uploadRes.json();
      const uploaded = data[0];

      // 2. Call dedicated /api/compress service
      const compressReq = {
        id: uploaded.id,
        format: targetFormat,
        quality: quality,
      };

      const res = await fetch(`${BACKEND_URL}/api/compress`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(compressReq),
      });

      if (!res.ok) throw new Error(await res.text());
      const blob = await res.blob();
      const downloadUrl = URL.createObjectURL(blob);

      const a = document.createElement('a');
      a.href = downloadUrl;
      a.download = `compressed-${Date.now()}.${targetFormat}`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
    } catch (err) {
      alert(`Compression failed: ${err.message}`);
    } finally {
      setIsProcessing(false);
    }
  };

  return (
    <div style={{ display: 'grid', gridTemplateColumns: '320px 1fr', gap: '20px', height: '100%' }}>
      <div className="sidebar glass-panel">
        <h2>Image Compressor</h2>

        <div 
          className="upload-zone" 
          onClick={() => fileInputRef.current?.click()}
          style={{ minHeight: '160px', justifyContent: 'center' }}
        >
          <input 
            type="file" 
            ref={fileInputRef} 
            accept="image/*" 
            style={{ display: 'none' }}
            onChange={handleFileSelect} 
          />
          <svg className="upload-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth="1.5">
            <path strokeLinecap="round" strokeLinejoin="round" d="M19.5 14.25v-2.625a3.375 3.375 0 00-3.375-3.375h-1.5A1.125 1.125 0 0113.5 7.125v-1.5a3.375 3.375 0 00-3.375-3.375H8.25m6.75 12l-3-3m0 0l-3 3m3-3v6m-1.5-15H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 00-9-9z" />
          </svg>
          <p style={{ fontSize: '13px', fontWeight: '500' }}>
            {selectedFile ? selectedFile.name : 'Select Single Image'}
          </p>
        </div>

        <div className="control-group" style={{ marginTop: '16px' }}>
          <label>Target Format</label>
          <select 
            className="control-select"
            value={targetFormat}
            onChange={(e) => setTargetFormat(e.target.value)}
          >
            <option value="webp">WebP (Modern & Compact)</option>
            <option value="jpeg">JPEG (High Compatibility)</option>
            <option value="png">PNG (Lossless Quality)</option>
          </select>
        </div>

        <div className="control-group">
          <label>Quality ({quality}%)</label>
          <input 
            type="range" 
            min="10" 
            max="100" 
            value={quality}
            onChange={(e) => setQuality(Number(e.target.value))}
            style={{ accentColor: 'var(--primary)', cursor: 'pointer' }}
          />
        </div>

        <button 
          className="btn-primary action-btn" 
          disabled={!selectedFile || isProcessing}
          onClick={handleCompress}
        >
          {isProcessing ? 'Compressing...' : 'Compress & Download'}
        </button>
      </div>

      <div className="viewport-container glass-panel" style={{ display: 'flex', flexDirection: 'column', justifyContent: 'center' }}>
        {previewUrl ? (
          <img 
            src={previewUrl} 
            alt="Target preview" 
            style={{ maxWidth: '90%', maxHeight: '80%', borderRadius: '12px', objectFit: 'contain', boxShadow: 'var(--shadow-lg)' }} 
          />
        ) : (
          <div style={{ textAlign: 'center', color: 'var(--text-muted)' }}>
            <p>Select an image to preview compression options</p>
          </div>
        )}
      </div>
    </div>
  );
}
