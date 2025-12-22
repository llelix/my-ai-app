import React, { useState, useEffect, useCallback } from 'react';
import { Button, Progress, Tooltip, Space, message, Modal } from 'antd';
import { 
  PlayCircleOutlined, 
  ThunderboltOutlined, 
  LoadingOutlined, 
  CheckCircleOutlined, 
  ExclamationCircleOutlined,
  ReloadOutlined,
  StopOutlined
} from '@ant-design/icons';
import { documentProcessingService } from '../services';
import type { DocumentProcessingStatus, ProcessingStatusType } from '../types';

interface DocumentProcessingButtonsProps {
  documentId: string;
  initialStatus?: DocumentProcessingStatus;
  onStatusChange?: (status: DocumentProcessingStatus) => void;
  disabled?: boolean;
  size?: 'small' | 'middle' | 'large';
  showProgress?: boolean;
  showRetry?: boolean;
}

/**
 * 文档处理操作按钮组件
 * 提供预处理和向量化操作的按钮，包含状态显示和进度指示器
 */
export const DocumentProcessingButtons: React.FC<DocumentProcessingButtonsProps> = ({
  documentId,
  initialStatus,
  onStatusChange,
  disabled = false,
  size = 'small',
  showProgress = true,
  showRetry = true
}) => {
  const [status, setStatus] = useState<DocumentProcessingStatus | null>(initialStatus || null);
  const [loading, setLoading] = useState<{
    preprocess: boolean;
    vectorize: boolean;
  }>({ preprocess: false, vectorize: false });

  // 状态更新处理
  const handleStatusUpdate = useCallback((newStatus: DocumentProcessingStatus) => {
    setStatus(newStatus);
    onStatusChange?.(newStatus);
  }, [onStatusChange]);

  // 获取初始状态
  useEffect(() => {
    if (!initialStatus && documentId) {
      documentProcessingService.getProcessingStatus(documentId)
        .then(handleStatusUpdate)
        .catch(error => {
          console.error('Failed to get initial processing status:', error);
          // 设置默认状态
          handleStatusUpdate({
            documentId,
            preprocessStatus: 'not_started',
            vectorizationStatus: 'not_started',
            progress: 0,
            vectorizationProgress: 0,
            createdAt: new Date().toISOString(),
            updatedAt: new Date().toISOString()
          });
        });
    }
  }, [documentId, initialStatus, handleStatusUpdate]);

  // 组件卸载时清理轮询
  useEffect(() => {
    return () => {
      documentProcessingService.stopPolling(documentId);
    };
  }, [documentId]);

  // 触发预处理
  const handlePreprocess = async () => {
    try {
      setLoading(prev => ({ ...prev, preprocess: true }));
      
      const newStatus = await documentProcessingService.triggerPreprocessing(documentId);
      handleStatusUpdate(newStatus);
      
      message.success('预处理已开始');
      
      // 开始轮询状态
      documentProcessingService.startPolling(documentId, handleStatusUpdate, {
        stopOnComplete: true
      });
      
    } catch (error) {
      console.error('Failed to trigger preprocessing:', error);
      message.error('启动预处理失败');
    } finally {
      setLoading(prev => ({ ...prev, preprocess: false }));
    }
  };

  // 触发向量化
  const handleVectorize = async () => {
    try {
      setLoading(prev => ({ ...prev, vectorize: true }));
      
      const newStatus = await documentProcessingService.triggerVectorization(documentId);
      handleStatusUpdate(newStatus);
      
      message.success('向量化已开始');
      
      // 开始轮询状态
      documentProcessingService.startPolling(documentId, handleStatusUpdate, {
        stopOnComplete: true
      });
      
    } catch (error) {
      console.error('Failed to trigger vectorization:', error);
      message.error('启动向量化失败');
    } finally {
      setLoading(prev => ({ ...prev, vectorize: false }));
    }
  };

  // 重试处理
  const handleRetry = (processType: 'preprocess' | 'vectorize') => {
    Modal.confirm({
      title: `重试${processType === 'preprocess' ? '预处理' : '向量化'}`,
      content: `确定要重试${processType === 'preprocess' ? '预处理' : '向量化'}操作吗？`,
      onOk: async () => {
        try {
          const newStatus = await documentProcessingService.retryProcessing(documentId, processType);
          handleStatusUpdate(newStatus);
          message.success(`${processType === 'preprocess' ? '预处理' : '向量化'}重试已开始`);
          
          // 开始轮询状态
          documentProcessingService.startPolling(documentId, handleStatusUpdate, {
            stopOnComplete: true
          });
        } catch (error) {
          console.error(`Failed to retry ${processType}:`, error);
          message.error(`重试${processType === 'preprocess' ? '预处理' : '向量化'}失败`);
        }
      }
    });
  };

  // 取消处理
  const handleCancel = (processType: 'preprocess' | 'vectorize') => {
    Modal.confirm({
      title: `取消${processType === 'preprocess' ? '预处理' : '向量化'}`,
      content: `确定要取消正在进行的${processType === 'preprocess' ? '预处理' : '向量化'}操作吗？`,
      onOk: async () => {
        try {
          await documentProcessingService.cancelProcessing(documentId, processType);
          message.success(`${processType === 'preprocess' ? '预处理' : '向量化'}已取消`);
        } catch (error) {
          console.error(`Failed to cancel ${processType}:`, error);
          message.error(`取消${processType === 'preprocess' ? '预处理' : '向量化'}失败`);
        }
      }
    });
  };

  // 获取按钮图标
  const getButtonIcon = (statusType: ProcessingStatusType, isLoading: boolean) => {
    if (isLoading) return <LoadingOutlined />;
    
    switch (statusType) {
      case 'not_started':
        return <PlayCircleOutlined />;
      case 'pending':
      case 'processing':
        return <LoadingOutlined />;
      case 'completed':
        return <CheckCircleOutlined />;
      case 'failed':
        return <ExclamationCircleOutlined />;
      default:
        return <PlayCircleOutlined />;
    }
  };

  // 获取按钮类型
  const getButtonType = (statusType: ProcessingStatusType): 'default' | 'primary' | 'dashed' => {
    switch (statusType) {
      case 'not_started':
        return 'primary';
      case 'pending':
      case 'processing':
        return 'default';
      case 'completed':
        return 'dashed';
      case 'failed':
        return 'default';
      default:
        return 'default';
    }
  };

  // 检查按钮是否应该禁用
  const isButtonDisabled = (statusType: ProcessingStatusType, isLoading: boolean) => {
    return disabled || isLoading || documentProcessingService.isProcessing(statusType);
  };

  if (!status) {
    return (
      <Space size="small">
        <Button size={size} loading>
          加载中...
        </Button>
      </Space>
    );
  }

  return (
    <Space direction="vertical" size="small" style={{ width: '100%' }}>
      <Space size="small" wrap>
        {/* 预处理按钮 */}
        <Tooltip title={`预处理状态: ${documentProcessingService.getStatusText(status.preprocessStatus)}`}>
          <Button
            size={size}
            type={getButtonType(status.preprocessStatus)}
            icon={getButtonIcon(status.preprocessStatus, loading.preprocess)}
            loading={loading.preprocess}
            disabled={isButtonDisabled(status.preprocessStatus, loading.preprocess)}
            onClick={handlePreprocess}
            style={{ minWidth: size === 'small' ? '70px' : '80px' }}
          >
            <span className="button-text">预处理</span>
          </Button>
        </Tooltip>

        {/* 向量化按钮 */}
        <Tooltip title={`向量化状态: ${documentProcessingService.getStatusText(status.vectorizationStatus)}`}>
          <Button
            size={size}
            type={getButtonType(status.vectorizationStatus)}
            icon={<ThunderboltOutlined />}
            loading={loading.vectorize}
            disabled={isButtonDisabled(status.vectorizationStatus, loading.vectorize)}
            onClick={handleVectorize}
            style={{ minWidth: size === 'small' ? '70px' : '80px' }}
          >
            <span className="button-text">向量化</span>
          </Button>
        </Tooltip>

        {/* 重试按钮 */}
        {showRetry && (
          <>
            {documentProcessingService.isFailed(status.preprocessStatus) && (
              <Tooltip title="重试预处理">
                <Button
                  size={size}
                  icon={<ReloadOutlined />}
                  onClick={() => handleRetry('preprocess')}
                  style={{ minWidth: size === 'small' ? '50px' : '60px' }}
                >
                  <span className="button-text">重试</span>
                </Button>
              </Tooltip>
            )}
            
            {documentProcessingService.isFailed(status.vectorizationStatus) && (
              <Tooltip title="重试向量化">
                <Button
                  size={size}
                  icon={<ReloadOutlined />}
                  onClick={() => handleRetry('vectorize')}
                  style={{ minWidth: size === 'small' ? '50px' : '60px' }}
                >
                  <span className="button-text">重试</span>
                </Button>
              </Tooltip>
            )}
          </>
        )}

        {/* 取消按钮 */}
        {(documentProcessingService.isProcessing(status.preprocessStatus) || 
          documentProcessingService.isProcessing(status.vectorizationStatus)) && (
          <Tooltip title="取消处理">
            <Button
              size={size}
              danger
              icon={<StopOutlined />}
              onClick={() => {
                if (documentProcessingService.isProcessing(status.preprocessStatus)) {
                  handleCancel('preprocess');
                } else if (documentProcessingService.isProcessing(status.vectorizationStatus)) {
                  handleCancel('vectorize');
                }
              }}
              style={{ minWidth: size === 'small' ? '50px' : '60px' }}
            >
              <span className="button-text">取消</span>
            </Button>
          </Tooltip>
        )}
      </Space>

      {/* 进度条 */}
      {showProgress && (
        <Space direction="vertical" size="small" style={{ width: '100%' }}>
          {(documentProcessingService.isProcessing(status.preprocessStatus) || 
            documentProcessingService.isCompleted(status.preprocessStatus)) && (
            <div>
              <div style={{ fontSize: '12px', marginBottom: '4px' }}>
                预处理进度: {Math.round(status.progress)}%
              </div>
              <Progress
                size="small"
                percent={Math.round(status.progress)}
                status={documentProcessingService.isFailed(status.preprocessStatus) ? 'exception' : 'active'}
                showInfo={false}
              />
            </div>
          )}
          
          {(documentProcessingService.isProcessing(status.vectorizationStatus) || 
            documentProcessingService.isCompleted(status.vectorizationStatus)) && (
            <div>
              <div style={{ fontSize: '12px', marginBottom: '4px' }}>
                向量化进度: {Math.round(status.vectorizationProgress)}%
              </div>
              <Progress
                size="small"
                percent={Math.round(status.vectorizationProgress)}
                status={documentProcessingService.isFailed(status.vectorizationStatus) ? 'exception' : 'active'}
                showInfo={false}
              />
            </div>
          )}
        </Space>
      )}

      {/* 错误信息 */}
      {(status.error || status.vectorizationError) && (
        <div style={{ fontSize: '12px', color: '#ff4d4f' }}>
          {status.error && <div>预处理错误: {status.error}</div>}
          {status.vectorizationError && <div>向量化错误: {status.vectorizationError}</div>}
        </div>
      )}

      {/* 响应式样式 */}
      <style dangerouslySetInnerHTML={{
        __html: `
          @media (max-width: 768px) {
            .button-text {
              display: none;
            }
          }
          @media (max-width: 480px) {
            .ant-space-item {
              margin-bottom: 4px !important;
            }
          }
        `
      }} />
    </Space>
  );
};

export default DocumentProcessingButtons;