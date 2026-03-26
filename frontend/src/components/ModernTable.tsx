import React, { useState, useMemo, useRef, useCallback } from 'react';
import { Button, Input } from 'antd';
import { LeftOutlined, RightOutlined, SearchOutlined } from '@ant-design/icons';

interface Column {
  key: string;
  title: string;
  render?: (value: any, record: any) => React.ReactNode;
  width?: number | string;
  minWidth?: number;
  sortable?: boolean;
  filterable?: boolean;
}

interface ModernTableProps {
  columns: Column[];
  data: any[];
  pageSize?: number;
  loading?: boolean;
  emptyText?: string;
  height?: string | number;
}

const ModernTable: React.FC<ModernTableProps> = ({ 
  columns, 
  data, 
  pageSize = 50, 
  loading = false,
  emptyText = "No data available",
  height = "400px"
}) => {
  const [currentPage, setCurrentPage] = useState(1);
  const [sortConfig, setSortConfig] = useState<{ key: string; direction: 'asc' | 'desc' } | null>(null);
  const [columnFilters, setColumnFilters] = useState<Record<string, string>>({});
  const [columnWidths, setColumnWidths] = useState<Record<string, number>>({});
  const [isResizing, setIsResizing] = useState<string | null>(null);
  const tableRef = useRef<HTMLDivElement>(null);

  // Filter and sort data
  const filteredAndSortedData = useMemo(() => {
    let filtered = data;

    // Apply filters
    Object.entries(columnFilters).forEach(([key, filter]) => {
      if (filter.trim()) {
        filtered = filtered.filter(item => {
          const value = item[key];
          const stringValue = value ? String(value).toLowerCase() : '';
          return stringValue.includes(filter.toLowerCase());
        });
      }
    });

    // Apply sorting
    if (sortConfig) {
      filtered = [...filtered].sort((a, b) => {
        const aValue = a[sortConfig.key];
        const bValue = b[sortConfig.key];

        if (aValue < bValue) {
          return sortConfig.direction === 'asc' ? -1 : 1;
        }
        if (aValue > bValue) {
          return sortConfig.direction === 'asc' ? 1 : -1;
        }
        return 0;
      });
    }

    return filtered;
  }, [data, sortConfig, columnFilters]);

  // Paginate data
  const totalPages = Math.ceil(filteredAndSortedData.length / pageSize);
  const startIndex = (currentPage - 1) * pageSize;
  const paginatedData = filteredAndSortedData.slice(startIndex, startIndex + pageSize);

  const handleSort = (key: string) => {
    setSortConfig(current => {
      if (!current || current.key !== key) {
        return { key, direction: 'asc' };
      }
      if (current.direction === 'asc') {
        return { key, direction: 'desc' };
      }
      return null;
    });
  };

  const handleFilterChange = (key: string, value: string) => {
    setColumnFilters(prev => ({ ...prev, [key]: value }));
    setCurrentPage(1); // Reset to first page when filtering
  };

  const handleMouseDown = useCallback((e: React.MouseEvent, columnKey: string) => {
    e.preventDefault();
    setIsResizing(columnKey);
    
    const startX = e.clientX;
    const startWidth = columnWidths[columnKey] || 150;

    const handleMouseMove = (e: MouseEvent) => {
      const diff = e.clientX - startX;
      const newWidth = Math.max(50, startWidth + diff);
      setColumnWidths(prev => ({ ...prev, [columnKey]: newWidth }));
    };

    const handleMouseUp = () => {
      setIsResizing(null);
      document.removeEventListener('mousemove', handleMouseMove);
      document.removeEventListener('mouseup', handleMouseUp);
    };

    document.addEventListener('mousemove', handleMouseMove);
    document.addEventListener('mouseup', handleMouseUp);
  }, [columnWidths]);

  const getColumnWidth = (column: Column) => {
    if (columnWidths[column.key]) return columnWidths[column.key];
    if (typeof column.width === 'number') return column.width;
    if (typeof column.width === 'string' && column.width.endsWith('px')) {
      return parseInt(column.width);
    }
    return 150; // default width
  };

  if (loading) {
    return (
      <div style={{ 
        padding: '40px', 
        textAlign: 'center', 
        border: '1px solid var(--border-color)',
        borderRadius: '6px',
        backgroundColor: 'var(--card-bg)'
      }}>
        Loading...
      </div>
    );
  }

  if (data.length === 0) {
    return (
      <div style={{ 
        padding: '40px', 
        textAlign: 'center', 
        border: '1px solid var(--border-color)',
        borderRadius: '6px',
        backgroundColor: 'var(--card-bg)',
        color: 'var(--app-text)'
      }}>
        {emptyText}
      </div>
    );
  }

  return (
    <div 
      ref={tableRef}
      style={{ 
        border: '1px solid var(--border-color)', 
        borderRadius: '6px',
        backgroundColor: 'var(--card-bg)',
        overflow: 'hidden',
        height: height,
        display: 'flex',
        flexDirection: 'column'
      }}
    >
      {/* Table Container with sticky header */}
      <div style={{ 
        flex: 1,
        overflow: 'auto',
        position: 'relative'
      }}>
        <table style={{ 
          width: '100%', 
          borderCollapse: 'separate',
          borderSpacing: 0,
          fontSize: '0.875rem'
        }}>
          <thead style={{
            position: 'sticky',
            top: 0,
            zIndex: 10,
            backgroundColor: 'var(--table-stripe)'
          }}>
            <tr>
              {columns.map(column => (
                <th
                  key={column.key}
                  style={{
                    padding: 0,
                    textAlign: 'left',
                    fontWeight: 600,
                    color: 'var(--app-text)',
                    borderBottom: '2px solid var(--border-color)',
                    width: getColumnWidth(column),
                    minWidth: column.minWidth || 50,
                    position: 'relative',
                    backgroundColor: 'var(--table-stripe)'
                  }}
                >
                  {/* Header content */}
                  <div style={{ 
                    padding: '8px 12px',
                    display: 'flex',
                    flexDirection: 'column',
                    gap: '4px'
                  }}>
                    {/* Title and sort */}
                    <div 
                      style={{
                        display: 'flex', 
                        alignItems: 'center', 
                        gap: '4px',
                        cursor: column.sortable !== false ? 'pointer' : 'default',
                        userSelect: 'none'
                      }}
                      onClick={() => column.sortable !== false && handleSort(column.key)}
                    >
                      {column.title}
                      {column.sortable !== false && sortConfig?.key === column.key && (
                        <span style={{ fontSize: '12px' }}>
                          {sortConfig.direction === 'asc' ? '↑' : '↓'}
                        </span>
                      )}
                    </div>
                    
                    {/* Filter input */}
                    {column.filterable !== false && (
                      <Input
                        size="small"
                        placeholder="Filter..."
                        prefix={<SearchOutlined style={{ fontSize: '12px' }} />}
                        value={columnFilters[column.key] || ''}
                        onChange={(e) => handleFilterChange(column.key, e.target.value)}
                        onClick={(e) => e.stopPropagation()}
                        style={{ fontSize: '12px' }}
                      />
                    )}
                  </div>
                  
                  {/* Resize handle */}
                  <div
                    style={{
                      position: 'absolute',
                      right: 0,
                      top: 0,
                      bottom: 0,
                      width: '4px',
                      cursor: 'col-resize',
                      backgroundColor: isResizing === column.key ? '#1890ff' : 'transparent',
                      borderRight: isResizing === column.key ? '2px solid #1890ff' : '1px solid var(--border-color)',
                      opacity: isResizing === column.key ? 1 : 0.3
                    }}
                    onMouseDown={(e) => handleMouseDown(e, column.key)}
                    onMouseEnter={(e) => {
                      if (!isResizing) {
                        e.currentTarget.style.opacity = '0.7';
                        e.currentTarget.style.borderRight = '1px solid var(--app-text)';
                      }
                    }}
                    onMouseLeave={(e) => {
                      if (!isResizing) {
                        e.currentTarget.style.opacity = '0.3';
                        e.currentTarget.style.borderRight = '1px solid var(--border-color)';
                      }
                    }}
                  />
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {paginatedData.map((record, index) => (
              <tr
                key={index}
                style={{ 
                  backgroundColor: index % 2 === 0 ? 'var(--card-bg)' : 'var(--table-stripe)',
                  borderBottom: '1px solid var(--border-color)'
                }}
                onMouseEnter={(e) => {
                  e.currentTarget.style.backgroundColor = 'var(--table-hover)';
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.style.backgroundColor = index % 2 === 0 ? 'var(--card-bg)' : 'var(--table-stripe)';
                }}
              >
                {columns.map(column => (
                  <td
                    key={column.key}
                    style={{
                      padding: '12px',
                      color: 'var(--app-text)',
                      fontFamily: 'monospace',
                      fontSize: '0.85rem',
                      width: getColumnWidth(column),
                      maxWidth: getColumnWidth(column),
                      overflow: 'hidden',
                      textOverflow: 'ellipsis'
                    }}
                  >
                    {column.render 
                      ? column.render(record[column.key], record)
                      : record[column.key]
                    }
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Pagination */}
      <div style={{ 
        padding: '12px 16px',
        borderTop: '1px solid var(--border-color)',
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
        backgroundColor: 'var(--table-stripe)',
        minHeight: '50px'
      }}>
        <div style={{ color: 'var(--app-text)', fontSize: '0.875rem' }}>
          Showing {startIndex + 1}-{Math.min(startIndex + pageSize, filteredAndSortedData.length)} of {filteredAndSortedData.length} items
          {Object.keys(columnFilters).some(key => columnFilters[key]?.trim()) && (
            <span style={{ color: 'var(--sl-color-neutral-600)', marginLeft: '8px' }}>
              (filtered from {data.length} total)
            </span>
          )}
        </div>
        {totalPages > 1 && (
          <div style={{ display: 'flex', gap: '8px', alignItems: 'center' }}>
            <Button
              size="small"
              icon={<LeftOutlined />}
              disabled={currentPage === 1}
              onClick={() => setCurrentPage(currentPage - 1)}
            >
              Previous
            </Button>
            <span style={{ color: 'var(--app-text)', fontSize: '0.875rem', padding: '0 8px' }}>
              Page {currentPage} of {totalPages}
            </span>
            <Button
              size="small"
              icon={<RightOutlined />}
              disabled={currentPage === totalPages}
              onClick={() => setCurrentPage(currentPage + 1)}
            >
              Next
            </Button>
          </div>
        )}
      </div>
    </div>
  );
};

export default ModernTable;