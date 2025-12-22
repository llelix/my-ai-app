import React from 'react';
import { Tag, Tooltip, Progress, Space } from 'antd';
import { 
  ClockCircleOutlined, 
  LoadingOutlined, 
  CheckCircleOutlined, 
  ExclamationCircleOutlined,
  MinusCircleOutlined
} from '@ant-design/icons';
import { documentProcessingService } from '../services';
import type { DocumentProcessingStatus, ProcessingStatusType } from '../types';

interface DocumentProcessingStatusProps {
  status: DocumentProcessingStatus;
  showProgress?: boolean;
  showBothStatuses?: boolean;
  size?: 'small' | 'default';
}

/**
 * 文档处理状态显示组件
 * 用于在列表或表格中显示文档的处理状态
 */
export const DocumentProcessingStatusIndicator: React.FC<DocumentProcessingStatusProps> = ({
  status,
  showProgress = false,
  showBothStatuses = true,
  size = 'small'
}) => {
  // 获取状态图标
  const getStatusIcon = (statusType: ProcessingStatusType) => {
    switch (statusType) {
      case 'not_started':
        return <MinusCircleOutlined />;
      case 'pending':
        return <ClockCircleOutlined />;
      case 'processing':
        return <LoadingOutlined />;
      case 'completed':
        return <CheckCircleOutlined />;
      case 'failed':
        return <ExclamationCircleOutlined />;
      default:
        return <MinusCircleOutlined />;
    }
  };

  // 渲染单个状态标签
  const renderStatusTag = (
    statusType: ProcessingStatusType,
    label: string,
    progress?: number,
    error?: string
  ) => {
    const color = documentProcessingService.getStatusColor(statusType);
    const text = documentProcessingService.getStatusText(statusType);
    const icon = getStatusIcon(statusType);

    const tooltipContent = (
      <div>
        <div>{label}: {text}</div>
        {progress !== undefined && progress > 0 && (
          <div>进度: {Math.round(progress)}%</div>
        )}
        {error && (
          <div style={{ color: '#ff4d4f' }}>错误: {error}</div>
        )}
        {status.processingTime && (
          <div>处理时间: {Math.round(status.processingTime / 1000)}秒</div>
        )}
      </div>
    );

    return (
      <Tooltip title={tooltipContent} key={label}>
        <Tag 
          color={color} 
          icon={icon}
          style={{ 
            margin: '2px',
            fontSize: size === 'small' ? '11px' : '12px'
          }}
        >
          {label}
          {showProgress && progress !== undefined && progress > 0 && (
            <span style={{ marginLeft: '4px' }}>
              ({Math.round(progress)}%)
            </span>
          )}
        </Tag>
      </Tooltip>
    );
  };

  return (
    <Space direction="vertical" size="small">
      <Space size="small" wrap>
        {/* 预处理状态 */}
        {renderStatusTag(
          status.preprocessStatus,
          '预处理',
          status.progress,
          status.error
        )}

        {/* 向量化状态 */}
        {showBothStatuses && renderStatusTag(
          status.vectorizationStatus,
          '向量化',
          status.vectorizationProgress,
          status.vectorizationError
        )}
      </Space>

      {/* 详细进度条 */}
      {showProgress && (
        <Space direction="vertical" size="small" style={{ width: '100%' }}>
          {(documentProcessingService.isProcessing(status.preprocessStatus) || 
            documentProcessingService.isCompleted(status.preprocessStatus)) && 
            status.progress > 0 && (
            <div>
              <div style={{ fontSize: '11px', marginBottom: '2px' }}>
                预处理: {Math.round(status.progress)}%
              </div>
              <Progress
                size="small"
                percent={Math.round(status.progress)}
                status={documentProcessingService.isFailed(status.preprocessStatus) ? 'exception' : 'active'}
                showInfo={false}
                strokeWidth={4}
              />
            </div>
          )}
          
          {showBothStatuses && 
           (documentProcessingService.isProcessing(status.vectorizationStatus) || 
            documentProcessingService.isCompleted(status.vectorizationStatus)) && 
            status.vectorizationProgress > 0 && (
            <div>
              <div style={{ fontSize: '11px', marginBottom: '2px' }}>
                向量化: {Math.round(status.vectorizationProgress)}%
              </div>
              <Progress
                size="small"
                percent={Math.round(status.vectorizationProgress)}
                status={documentProcessingService.isFailed(status.vectorizationStatus) ? 'exception' : 'active'}
                showInfo={false}
                strokeWidth={4}
              />
            </div>
          )}
        </Space>
      )}
    </Space>
  );
};

export default DocumentProcessingStatusIndicator;