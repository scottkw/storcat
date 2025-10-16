import React from 'react';
import { Card, Typography } from 'antd';

const { Title, Paragraph } = Typography;

function WelcomeContent() {
  return (
    <Card
      style={{
        background: 'var(--card-bg)',
        border: '1px solid var(--border-color)',
        borderRadius: 8,
        marginBottom: 24,
        boxShadow: '0 2px 8px var(--shadow-color)',
      }}
    >
      <Title level={2} style={{ margin: '0 0 16px 0', color: 'var(--app-text)' }}>
        Welcome to StorCat
      </Title>
      <Paragraph style={{ marginBottom: 24, color: 'var(--app-text)' }}>
        Use the tabs above to create new directory catalogs, search existing catalogs, or browse your catalog collection.
      </Paragraph>
      
      <Title level={3} style={{ margin: '24px 0 16px 0', fontSize: '1.1rem', fontWeight: 600, color: 'var(--app-text)' }}>
        Features
      </Title>
      <ul style={{ margin: 0, paddingLeft: 20, lineHeight: 1.6, color: 'var(--app-text)' }}>
        <li><strong>Create Catalogs:</strong> Generate JSON and HTML catalogs of any directory</li>
        <li><strong>Search Catalogs:</strong> Find files across multiple catalog files</li>
        <li><strong>Browse Catalogs:</strong> View and manage your catalog collection</li>
        <li><strong>Cross-Platform:</strong> Pure JavaScript implementation works on all platforms</li>
        <li><strong>Compatible:</strong> 100% compatible with existing catalog files</li>
      </ul>
    </Card>
  );
}

export default WelcomeContent;