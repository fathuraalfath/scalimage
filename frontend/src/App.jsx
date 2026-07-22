import React, { useState, useRef } from 'react';
import { PAPER_SIZES, TEMPLATES } from './helpers/layout';
import './App.css';

const BACKEND_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

function App() {
  const [uploadedImages, setUploadedImages] = useState([]);
  const [selectedPaperSize, setSelectedPaperSize] = useState('A4');
  const [orientation, setOrientation] = useState('portrait');
  const [bgColor, setBgColor] = useState('#1e293b');
  const [exportFormat, setExportFormat] = useState('png');
  const [activeTemplateIndex, setActiveTemplateIndex] = useState(0);
  const [isUploading, setIsUploading] = useState(false);
  const [isGenerating, setIsGenerating] = useState(false);
  const [isDragging, setIsDragging] = useState(false);
  const fileInputRef = useRef(null);

  // Compute paper size dimensions
  const paper = PAPER_SIZES[selectedPaperSize];
  const canvasWidth = orientation === 'portrait' ? paper.width : paper.height;
  const canvasHeight = orientation === 'portrait' ? paper.height : paper.width;

  // Visual scaling to fit viewport nicely
  const maxViewportWidth = 650;
  const maxViewportHeight = 600;
  const scale = Math.min(maxViewportWidth / canvasWidth, maxViewportHeight / canvasHeight, 1);
  const displayWidth = Math.round(canvasWidth * scale);
  const displayHeight = Math.round(canvasHeight * scale);

  // Determine slot templates based on the current count of images
  const slotCount = Math.min(Math.max(uploadedImages.length, 1), 8);
  const availableTemplates = TEMPLATES[slotCount] || TEMPLATES[1];
  
  // Safe bounds check for active template index
  const activeTemplate = availableTemplates[activeTemplateIndex] 
    ? availableTemplates[activeTemplateIndex] 
    : availableTemplates[0];

  // Upload handler
  const handleFiles = async (files) => {
    if (!files || files.length === 0) return;

    // Phase 1 Limit: check total image count
    if (uploadedImages.length + files.length > 8) {
      alert('Phase 1 Collage Maker supports a maximum of 8 images. Please reduce the number of files.');
      return;
    }

    setIsUploading(true);
    const formData = new FormData();
    for (let i = 0; i < files.length; i++) {
      formData.append('images', files[i]);
    }

    try {
      const response = await fetch(`${BACKEND_URL}/api/upload`, {
        method: 'POST',
        body: formData,
      });

      if (!response.ok) {
        throw new Error(await response.text());
      }

      const data = await response.json();
      
      // Merge with existing uploads
      setUploadedImages((prev) => [
        ...prev,
        ...data.map((img, idx) => ({
          id: img.id,
          width: img.width,
          height: img.height,
          format: img.format,
          name: files[idx]?.name || `Image-${img.id.slice(0, 5)}`
        }))
      ]);
      
      // Reset template selection when count changes
      setActiveTemplateIndex(0);
    } catch (err) {
      console.error(err);
      alert(`Upload failed: ${err.message}`);
    } finally {
      setIsUploading(false);
    }
  };

  const handleDrop = (e) => {
    e.preventDefault();
    setIsDragging(false);
    if (e.dataTransfer.files) {
      handleFiles(e.dataTransfer.files);
    }
  };

  const handleDragOver = (e) => {
    e.preventDefault();
    setIsDragging(true);
  };

  const handleDragLeave = () => {
    setIsDragging(false);
  };

  const removeImage = (index) => {
    setUploadedImages((prev) => prev.filter((_, idx) => idx !== index));
    setActiveTemplateIndex(0);
  };

  // Reordering helpers to swap slot layout assignment
  const moveImage = (index, direction) => {
    const nextIndex = index + direction;
    if (nextIndex < 0 || nextIndex >= uploadedImages.length) return;

    setUploadedImages((prev) => {
      const updated = [...prev];
      const temp = updated[index];
      updated[index] = updated[nextIndex];
      updated[nextIndex] = temp;
      return updated;
    });
  };

  // Trigger collage compilation
  const generateCollage = async () => {
    if (uploadedImages.length === 0) {
      alert('Please upload at least one image to compile.');
      return;
    }

    setIsGenerating(true);

    // Map percentage layout slots to actual canvas resolution pixels
    const resolvedImages = activeTemplate.slots.map((slot, index) => {
      const img = uploadedImages[index];
      return {
        id: img.id,
        x: Math.round(slot.x * canvasWidth),
        y: Math.round(slot.y * canvasHeight),
        width: Math.round(slot.w * canvasWidth),
        height: Math.round(slot.h * canvasHeight),
      };
    });

    const payload = {
      canvasWidth,
      canvasHeight,
      bgColor,
      format: exportFormat,
      images: resolvedImages,
    };

    try {
      const response = await fetch(`${BACKEND_URL}/api/collage`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(payload),
      });

      if (!response.ok) {
        throw new Error(await response.text());
      }

      const blob = await response.blob();
      const url = URL.createObjectURL(blob);

      // Download anchor trigger
      const link = document.createElement('a');
      link.href = url;
      link.download = `collage-${selectedPaperSize}-${orientation}-${Date.now()}.${exportFormat}`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      URL.revokeObjectURL(url);
    } catch (err) {
      console.error(err);
      alert(`Collage generation failed: ${err.message}`);
    } finally {
      setIsGenerating(false);
    }
  };

  return (
    <div className="app-container">
      {/* Top Header */}
      <header className="header glass-panel">
        <div className="header-title">
          <h1>
            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round" style={{color: 'var(--primary)'}}>
              <rect x="3" y="3" width="18" height="18" rx="2" ry="2"/>
              <circle cx="9" cy="9" r="2"/>
              <path d="m21 15-3.086-3.086a2 2 0 0 0-2.828 0L6 21"/>
            </svg>
            Scalimage
          </h1>
          <p>High-performance image collage generation platform</p>
        </div>
      </header>

      {/* Main Workspace Layout */}
      <main className="main-content">
        {/* Left Side: Upload & Media Manager */}
        <section className="sidebar glass-panel">
          <h2>1. Upload Images</h2>
          
          <div 
            className={`upload-zone ${isDragging ? 'dragging' : ''}`}
            onDragOver={handleDragOver}
            onDragLeave={handleDragLeave}
            onDrop={handleDrop}
            onClick={() => fileInputRef.current?.click()}
          >
            <input 
              type="file" 
              ref={fileInputRef} 
              multiple 
              accept="image/*" 
              style={{ display: 'none' }}
              onChange={(e) => handleFiles(e.target.files)}
            />
            <svg className="upload-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth="1.5">
              <path strokeLinecap="round" strokeLinejoin="round" d="M12 9v6m3-3H9m12 0a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            <div>
              <p style={{ fontWeight: '500', fontSize: '13px' }}>Click or Drag & Drop</p>
              <p style={{ fontSize: '11px', color: 'var(--text-muted)', marginTop: '4px' }}>PNG, JPG or WebP (Max 8)</p>
            </div>
          </div>

          {isUploading && (
            <div style={{ display: 'flex', alignItems: 'center', gap: '8px', fontSize: '12px', color: 'var(--accent)' }}>
              <div className="spinner" style={{ width: '16px', height: '16px', borderWidth: '2px' }}></div>
              Uploading files...
            </div>
          )}

          <h2 style={{ marginTop: '12px' }}>Media Files ({uploadedImages.length}/8)</h2>
          <div className="image-list">
            {uploadedImages.map((img, idx) => (
              <div key={img.id} className="image-thumb">
                <img src={`${BACKEND_URL}/uploads/${img.id}`} alt={img.name} />
                <button className="remove-btn" onClick={() => removeImage(idx)}>×</button>
                
                {/* Layer arrange overrides */}
                <div style={{
                  position: 'absolute',
                  bottom: '4px',
                  left: '4px',
                  display: 'flex',
                  gap: '2px',
                  background: 'rgba(0,0,0,0.7)',
                  borderRadius: '4px',
                  padding: '2px'
                }}>
                  <button 
                    disabled={idx === 0}
                    onClick={(e) => { e.stopPropagation(); moveImage(idx, -1); }}
                    style={{ background: 'none', border: 'none', color: '#fff', fontSize: '10px', cursor: 'pointer', opacity: idx === 0 ? 0.3 : 1 }}
                  >
                    ▲
                  </button>
                  <button 
                    disabled={idx === uploadedImages.length - 1}
                    onClick={(e) => { e.stopPropagation(); moveImage(idx, 1); }}
                    style={{ background: 'none', border: 'none', color: '#fff', fontSize: '10px', cursor: 'pointer', opacity: idx === uploadedImages.length - 1 ? 0.3 : 1 }}
                  >
                    ▼
                  </button>
                </div>
                
                <span className="slot-index">{idx + 1}</span>
              </div>
            ))}
          </div>
        </section>

        {/* Middle Area: Interactive Canvas Viewport */}
        <section className="viewport-container">
          {uploadedImages.length === 0 ? (
            <div className="empty-canvas-overlay">
              <svg fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth="1">
                <path strokeLinecap="round" strokeLinejoin="round" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
              </svg>
              <p style={{ fontSize: '14px', fontWeight: '500' }}>Your collage preview will appear here</p>
              <p style={{ fontSize: '12px', color: 'var(--text-dim)' }}>Upload images in the sidebar to populate slots</p>
            </div>
          ) : (
            <div 
              className="canvas-sheet"
              style={{
                width: `${displayWidth}px`,
                height: `${displayHeight}px`,
                backgroundColor: bgColor
              }}
            >
              {activeTemplate.slots.map((slot, index) => {
                const img = uploadedImages[index];
                return (
                  <div 
                    key={index} 
                    className="canvas-slot"
                    style={{
                      left: `${slot.x * 100}%`,
                      top: `${slot.y * 100}%`,
                      width: `${slot.w * 100}%`,
                      height: `${slot.h * 100}%`
                    }}
                  >
                    {img ? (
                      <img src={`${BACKEND_URL}/uploads/${img.id}`} alt="Placed" />
                    ) : (
                      <div className="slot-placeholder">
                        <svg fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth="1.5">
                          <path strokeLinecap="round" strokeLinejoin="round" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
                        </svg>
                        <span>Slot {index + 1}</span>
                      </div>
                    )}
                    <span className="slot-index">{index + 1}</span>
                  </div>
                );
              })}
            </div>
          )}

          {isGenerating && (
            <div className="loading-indicator">
              <div className="spinner"></div>
              <p style={{ fontWeight: '500', color: '#fff' }}>Compiling high-resolution collage...</p>
            </div>
          )}
        </section>

        {/* Right Side: Collage Properties & Compile controls */}
        <section className="sidebar glass-panel">
          <h2>2. Layout Settings</h2>

          <div className="control-group">
            <label>Paper Size</label>
            <select 
              className="control-select"
              value={selectedPaperSize}
              onChange={(e) => setSelectedPaperSize(e.target.value)}
            >
              <option value="A4">A4 (210 x 297 mm)</option>
              <option value="F4">F4 / Folio (215 x 330 mm)</option>
              <option value="A5">A5 (148 x 210 mm)</option>
              <option value="Square">Square (1:1 Ratio)</option>
            </select>
          </div>

          <div className="control-group">
            <label>Orientation</label>
            <div className="orientation-toggle">
              <button 
                className={`toggle-btn ${orientation === 'portrait' ? 'active' : ''}`}
                onClick={() => setOrientation('portrait')}
              >
                Portrait
              </button>
              <button 
                className={`toggle-btn ${orientation === 'landscape' ? 'active' : ''}`}
                onClick={() => setOrientation('landscape')}
              >
                Landscape
              </button>
            </div>
          </div>

          <div className="control-group">
            <label>Background Color</label>
            <div className="color-picker-wrapper">
              <input 
                type="color" 
                className="color-preview"
                value={bgColor}
                onChange={(e) => setBgColor(e.target.value)}
              />
              <input 
                type="text" 
                className="control-input"
                style={{ fontFamily: 'monospace' }}
                value={bgColor.toUpperCase()}
                onChange={(e) => setBgColor(e.target.value)}
              />
            </div>
          </div>

          <h2 style={{ marginTop: '12px' }}>3. Select Template</h2>
          <div className="control-group">
            <label>Pre-defined Grids ({slotCount} slots)</label>
            <div className="templates-list">
              {availableTemplates.map((temp, idx) => (
                <div 
                  key={idx}
                  className={`template-item ${activeTemplateIndex === idx ? 'active' : ''}`}
                  onClick={() => setActiveTemplateIndex(idx)}
                >
                  <span>{temp.name}</span>
                  <div style={{ display: 'flex', gap: '2px' }}>
                    {temp.slots.map((_, sIdx) => (
                      <div 
                        key={sIdx}
                        style={{
                          width: '8px',
                          height: '8px',
                          border: '1px solid currentColor',
                          borderRadius: '1px'
                        }}
                      ></div>
                    ))}
                  </div>
                </div>
              ))}
            </div>
          </div>

          <h2 style={{ marginTop: '12px' }}>4. Export</h2>
          <div className="control-group">
            <label>Output Format</label>
            <select 
              className="control-select"
              value={exportFormat}
              onChange={(e) => setExportFormat(e.target.value)}
            >
              <option value="png">PNG (Lossless Quality)</option>
              <option value="jpeg">JPEG (High Performance)</option>
            </select>
          </div>

          <button 
            className="btn-primary action-btn"
            disabled={uploadedImages.length === 0 || isGenerating}
            onClick={generateCollage}
          >
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
              <path strokeLinecap="round" strokeLinejoin="round" d="M3 16.5v2.25A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75V16.5M16.5 12 12 16.5m0 0L7.5 12m4.5 4.5V3" />
            </svg>
            {isGenerating ? 'Compiling...' : 'Download Collage'}
          </button>
          
          <div style={{ fontSize: '11px', color: 'var(--text-dim)', textAlign: 'center', marginTop: '10px' }}>
            Internal Output: {canvasWidth} x {canvasHeight} px
          </div>
        </section>
      </main>
    </div>
  );
}

export default App;
