import React, { useState, useRef } from 'react';

const BACKEND_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

export default function ResizeTool() {
  const [selectedFile, setSelectedFile] = useState(null);
  const [previewUrl, setPreviewUrl] = useState(null);
  const [originalWidth, setOriginalWidth] = useState(0);
  const [originalHeight, setOriginalHeight] = useState(0);
  const [targetWidth, setTargetWidth] = useState(800);
  const [targetHeight, setTargetHeight] = useState(600);
  const [keepAspect, setKeepAspect] = useState(true);
  const [isProcessing, setIsProcessing] = useState(false);
  const fileInputRef = useRef(null);

  const handleFileSelect = (e) => {
    const file = e.target.files?.[0];
    if (!file) return;

    setSelectedFile(file);
    setPreviewUrl(URL.createObjectURL(file));

    // Load dimensions
    const img = new Image();
    img.onload = () => {
      setOriginalWidth(img.width);
      setOriginalHeight(img.height);
      setTargetWidth(img.width);
      setTargetHeight(img.height);
    };
    img.src = URL.createObjectURL(file);
  };

  const handleWidthChange = (val) => {
    const w = Number(val);
    setTargetWidth(w);
    if (keepAspect && originalWidth > 0) {
      const ratio = originalHeight / originalWidth;
      setTargetHeight(Math.round(w * ratio));
    }
  };

  const handleHeightChange = (val) => {
    const h = Number(val);
    setTargetHeight(h);
    if (keepAspect && originalHeight > 0) {
      const ratio = originalWidth / originalHeight;
      setTargetWidth(Math.round(h * ratio));
    }
  };

  const handleResize = async () => {
    if (!selectedFile) return;

    setIsProcessing(true);

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

      // Call dedicated /api/resize service
      const resizeReq = {
        id: uploaded.id,
        targetWidth: targetWidth,
        targetHeight: targetHeight,
        format: 'png',
      };

      const res = await fetch(`${BACKEND_URL}/api/resize`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(resizeReq),
      });

      if (!res.ok) throw new Error(await res.text());
      const blob = await res.blob();
      const downloadUrl = URL.createObjectURL(blob);

      const a = document.createElement('a');
      a.href = downloadUrl;
      a.download = `resized-${targetWidth}x${targetHeight}-${Date.now()}.png`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
    } catch (err) {
      alert(`Resize failed: ${err.message}`);
    } finally {
      setIsProcessing(false);
    }
  };

  return (
    <div style={{ display: 'grid', gridTemplateColumns: '320px 1fr', gap: '20px', height: '100%' }}>
      <div className="sidebar glass-panel">
        <h2>Image Resizer</h2>

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
            <path strokeLinecap="round" strokeLinejoin="round" d="M3.75 3.75v4.5m0-4.5h4.5m-4.5 0L9 9M3.75 20.25v-4.5m0 4.5h4.5m-4.5 0L9 15M20.25 3.75h-4.5m4.5 0v4.5m0-4.5L15 9m5.25 11.25h-4.5m4.5 0v-4.5m0 4.5L15 15" />
          </svg>
          <p style={{ fontSize: '13px', fontWeight: '500' }}>
            {selectedFile ? selectedFile.name : 'Select Image to Resize'}
          </p>
        </div>

        {originalWidth > 0 && (
          <div style={{ fontSize: '11px', color: 'var(--text-muted)', textAlign: 'center', marginTop: '6px' }}>
            Original: {originalWidth} x {originalHeight} px
          </div>
        )}

        <div className="control-group" style={{ marginTop: '16px' }}>
          <label>Target Width (px)</label>
          <input 
            type="number" 
            className="control-input"
            value={targetWidth}
            onChange={(e) => handleWidthChange(e.target.value)}
          />
        </div>

        <div className="control-group">
          <label>Target Height (px)</label>
          <input 
            type="number" 
            className="control-input"
            value={targetHeight}
            onChange={(e) => handleHeightChange(e.target.value)}
          />
        </div>

        <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '16px' }}>
          <input 
            type="checkbox" 
            id="aspect"
            checked={keepAspect}
            onChange={(e) => setKeepAspect(e.target.checked)}
            style={{ accentColor: 'var(--primary)' }}
          />
          <label htmlFor="aspect" style={{ fontSize: '12px', color: 'var(--text-muted)', cursor: 'pointer' }}>
            Maintain Aspect Ratio
          </label>
        </div>

        <button 
          className="btn-primary action-btn" 
          disabled={!selectedFile || isProcessing}
          onClick={handleResize}
        >
          {isProcessing ? 'Resizing...' : 'Resize & Download'}
        </button>
      </div>

      <div className="viewport-container glass-panel" style={{ display: 'flex', flexDirection: 'column', justifyContent: 'center' }}>
        {previewUrl ? (
          <div style={{ textAlign: 'center' }}>
            <img 
              src={previewUrl} 
              alt="Resize target preview" 
              style={{ maxWidth: '90%', maxHeight: '75%', borderRadius: '12px', objectFit: 'contain', boxShadow: 'var(--shadow-lg)' }} 
            />
            <p style={{ fontSize: '12px', color: 'var(--text-muted)', marginTop: '12px' }}>
              Target Resolution: {targetWidth} × {targetHeight} px
            </p>
          </div>
        ) : (
          <div style={{ textAlign: 'center', color: 'var(--text-muted)' }}>
            <p>Select an image to preview dimension scaling</p>
          </div>
        )}
      </div>
    </div>
  );
}
