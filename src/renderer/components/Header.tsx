import React from 'react';
import { Layout, Typography } from 'antd';
import storcatIcon from '../storcat-icon.svg';

const { Header: AntHeader } = Layout;
const { Title } = Typography;

function Header() {
  return (
    <AntHeader
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          padding: '0 24px',
          height: 'var(--header-height)',
          background: 'var(--header-bg)',
          borderBottom: 'none',
          // macOS traffic light button spacing
          WebkitAppRegion: 'drag' as any,
        }}
      >
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            flex: 1,
            paddingLeft: '70px', // Space for macOS window controls
            gap: '12px',
          }}
        >
          <img
            src={storcatIcon}
            alt="StorCat Icon"
            style={{
              width: '24px',
              height: '24px',
              filter: 'var(--icon-filter)',
            }}
          />
          <Title
            level={4}
            style={{
              margin: 0,
              color: 'var(--header-text)',
              fontSize: '1.25rem',
              fontWeight: 600,
            }}
          >
            StorCat - Directory Catalog Manager
          </Title>
        </div>
    </AntHeader>
  );
}

export default Header;