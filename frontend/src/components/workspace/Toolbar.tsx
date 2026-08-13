import React from 'react';

function Toolbar() {
  return (
    <div
      className="ws-toolbar"
      style={
        {
          '--wails-draggable': 'drag',
        } as React.CSSProperties & { '--wails-draggable'?: string }
      }
    >
      <span style={{ fontSize: '13px', fontWeight: 600, letterSpacing: '-0.01em' }}>StorCat</span>
    </div>
  );
}

export default Toolbar;
