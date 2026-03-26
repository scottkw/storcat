import React, { useState, useEffect } from 'react';
import { Modal, message } from 'antd';

interface CatalogModalProps {
  visible: boolean;
  catalogPath: string | null;
  onClose: () => void;
}

function CatalogModal({ visible, catalogPath, onClose }: CatalogModalProps) {
  const [htmlContent, setHtmlContent] = useState<string>('');
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (visible && catalogPath) {
      loadHtmlContent();
    }
  }, [visible, catalogPath]);

  const loadHtmlContent = async () => {
    if (!catalogPath) return;

    setLoading(true);
    try {
      // Get the HTML file path -- now returns {success, htmlPath} envelope
      const pathResult = await window.electronAPI.getCatalogHtmlPath(catalogPath);
      if (!pathResult.success) {
        message.error(pathResult.error || 'HTML file not found for this catalog');
        return;
      }

      // Read the HTML content -- now returns {success, content} envelope
      const readResult = await window.electronAPI.readHtmlFile(pathResult.htmlPath);
      if (!readResult.success) {
        message.error(readResult.error || 'Failed to read HTML file');
        return;
      }

      // Apply theme modifications to the HTML content
      const modifiedHtml = modifyHtmlForTheme(readResult.content);
      setHtmlContent(modifiedHtml);
    } catch (error) {
      message.error('Failed to load catalog HTML');
    } finally {
      setLoading(false);
    }
  };

  const modifyHtmlForTheme = (html: string): string => {
    // Check if we're in dark mode
    const isDark = document.documentElement.hasAttribute('data-theme');
    
    if (isDark) {
      // Apply dark theme styles by injecting CSS
      return html.replace(
        '<head>',
        `<head>
        <style>
          /* Dark mode overrides for catalog HTML */
          body { 
            background-color: #1e293b !important; 
            color: #f1f5f9 !important; 
          }
          /* Override any black text to white */
          * {
            color: #f1f5f9 !important;
          }
          /* Directory/folder styling */
          .directory, .folder { 
            background-color: #334155 !important; 
            color: #f1f5f9 !important; 
          }
          /* File styling */
          .file { 
            color: #cbd5e1 !important; 
          }
          /* Link styling */
          a { 
            color: #60a5fa !important; 
          }
          a:hover { 
            color: #93c5fd !important; 
          }
          /* Table styling if present */
          table, th, td {
            background-color: #334155 !important;
            color: #f1f5f9 !important;
            border-color: #475569 !important;
          }
          /* Pre/code blocks */
          pre, code {
            background-color: #334155 !important;
            color: #f1f5f9 !important;
          }
        </style>`
      );
    }
    
    // For light mode, return original HTML unchanged
    return html;
  };

  const handleClose = () => {
    setHtmlContent('');
    onClose();
  };

  return (
    <Modal
      title="Catalog HTML Preview"
      open={visible}
      onCancel={handleClose}
      footer={null}
      width="90%"
      style={{ top: 20 }}
      bodyStyle={{
        height: 'calc(100vh - 200px)',
        padding: 0,
        overflow: 'hidden'
      }}
    >
      {loading ? (
        <div style={{ 
          display: 'flex', 
          justifyContent: 'center', 
          alignItems: 'center', 
          height: '100%' 
        }}>
          Loading...
        </div>
      ) : (
        <iframe
          srcDoc={htmlContent}
          style={{
            width: '100%',
            height: '100%',
            border: 'none'
          }}
          title="Catalog HTML Content"
        />
      )}
    </Modal>
  );
}

export default CatalogModal;