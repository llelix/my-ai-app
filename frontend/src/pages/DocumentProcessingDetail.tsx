import React, { useState, useEffect, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
  Card,
  Descriptions,
  Progress,
  Button,
  Space,
  Timeline,
  Alert,
  Spin,
  Tag,
  Divider,
  Row,
  Col,
  Statistic,
  message,
  Modal
} from 'antd';
import {
  ArrowLeftOutlined,
  ReloadOutlined,
  PlayCircleOutlined,
  ThunderboltOutlined,
  ClockCircleOutlined,
  CheckCircleOutlined,
  ExclamationCircleOutlined,
  StopOutlined
} from '@ant-design/icons';
import { documentService, documentProcessingService } from '../services';
import type { Document, DocumentProcessingStatus } from '../types';

/**
 * 文档处理状态详情页面
 * 显示文档处理的详细状态、历史记录和操作选项
 */
const DocumentProcessingDetail: React.FC = () => {
  const { documentId } = useParams<{ documentId: string }>();
  const navigate = useNavigate();
  
  const [document, setDocument] = useState<Document | null>(null);
  const [processingStatus, setProcessingStatus] = useState<DocumentProcessingStatus | null>(null);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [autoRefresh] = useState(true);

  // 加载文档信息
  const loadDocument = useCallback(async () => {
    if (!documentId) return;
    
    try {
      const response = await documentService.getDocument(parseInt(documentId));
      if (response.data) {
        setDocument(response.data);
      }
    } catch (error) {
      console.error('Failed to load document:', error);
      message.error('加载文档信息失败');
    }
  }, [documentId]);

  // 加载处理状态
  const loadProcessingStatus = useCallback(async () => {
    if (!documentId) return;
    
    try {
      const status = await documentProcessingService.getProcessingStatus(documentId);
      setProcessingStatus(status);
    } catch (error) {
      console.error('Failed to load processing status:', error);
      message.error('加载处理状态失败');
    }
  }, [documentId]);

  // 刷新数据
  const refreshData = async () => {
    setRefreshing(true);
    try {
      await Promise.all([loadDocument(), loadProcessingStatus()]);
    } finally {
      setRefreshing(false);
    }
  };

  // 初始化数据加载
  useEffect(() => {
    const initializeData = async () => {
      setLoading(true);
      try {
        await Promise.all([loadDocument(), loadProcessingStatus()]);
      } finally {
        setLoading(false);
      }
    };

    if (documentId) {
      initializeData();
    }
  }, [documentId, loadDocument, loadProcessingStatus]);

  // 自动刷新逻辑
  useEffect(() => {
    if (!autoRefresh || !documentId || !processingStatus) return;

    // 如果有正在进行的处理，启动轮询
    const isProcessing = 
      documentProcessingService.isProcessing(processingStatus.preprocessStatus) ||
      documentProcessingService.isProcessing(processingStatus.vectorizationStatus);

    if (isProcessing) {
      documentProcessingService.startPolling(
        documentId,
        setProcessingStatus,
        { stopOnComplete: true }
      );
    }

    return () => {
      documentProcessingService.stopPolling(documentId);
    };
  }, [documentId, processingStatus, autoRefresh]);

  // 组件卸载时清理
  useEffect(() => {
    return () => {
      if (documentId) {
        documentProcessingService.stopPolling(documentId);
      }
    };
  }, [documentId]);

  // 触发预处理
  const handlePreprocess = async () => {
    if (!documentId) return;
    
    try {
      await documentProcessingService.triggerPreprocessing(documentId);
      message.success('预处理已开始');
      await refreshData();
    } catch (error) {
      console.error('Failed to trigger preprocessing:', error);
      message.error('启动预处理失败');
    }
  };

  // 触发向量化
  const handleVectorize = async () => {
    if (!documentId) return;
    
    try {
      await documentProcessingService.triggerVectorization(documentId);
      message.success('向量化已开始');
      await refreshData();
    } catch (error) {
      console.error('Failed to trigger vectorization:', error);
      message.error('启动向量化失败');
    }
  };

  // 重试处理
  const handleRetry = (processType: 'preprocess' | 'vectorize') => {
    if (!documentId) return;
    
    Modal.confirm({
      title: `重试${processType === 'preprocess' ? '预处理' : '向量化'}`,
      content: `确定要重试${processType === 'preprocess' ? '预处理' : '向量化'}操作吗？`,
      onOk: async () => {
        try {
          await documentProcessingService.retryProcessing(documentId, processType);
          message.success(`${processType === 'preprocess' ? '预处理' : '向量化'}重试已开始`);
          await refreshData();
        } catch (error) {
          console.error(`Failed to retry ${processType}:`, error);
          message.error(`重试${processType === 'preprocess' ? '预处理' : '向量化'}失败`);
        }
      }
    });
  };

  // 取消处理
  const handleCancel = (processType: 'preprocess' | 'vectorize') => {
    if (!documentId) return;
    
    Modal.confirm({
      title: `取消${processType === 'preprocess' ? '预处理' : '向量化'}`,
      content: `确定要取消正在进行的${processType === 'preprocess' ? '预处理' : '向量化'}操作吗？`,
      onOk: async () => {
        try {
          await documentProcessingService.cancelProcessing(documentId, processType);
          message.success(`${processType === 'preprocess' ? '预处理' : '向量化'}已取消`);
          await refreshData();
        } catch (error) {
          console.error(`Failed to cancel ${processType}:`, error);
          message.error(`取消${processType === 'preprocess' ? '预处理' : '向量化'}失败`);
        }
      }
    });
  };

  // 格式化文件大小
  const formatFileSize = (bytes: number) => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  // 格式化处理时间
  const formatProcessingTime = (milliseconds?: number) => {
    if (!milliseconds) return '-';
    const seconds = Math.floor(milliseconds / 1000);
    const minutes = Math.floor(seconds / 60);
    const hours = Math.floor(minutes / 60);
    
    if (hours > 0) {
      return `${hours}小时${minutes % 60}分钟${seconds % 60}秒`;
    } else if (minutes > 0) {
      return `${minutes}分钟${seconds % 60}秒`;
    } else {
      return `${seconds}秒`;
    }
  };

  // 获取状态图标
  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'not_started':
        return <ClockCircleOutlined style={{ color: '#d9d9d9' }} />;
      case 'pending':
        return <ClockCircleOutlined style={{ color: '#faad14' }} />;
      case 'processing':
        return <ReloadOutlined spin style={{ color: '#1890ff' }} />;
      case 'completed':
        return <CheckCircleOutlined style={{ color: '#52c41a' }} />;
      case 'failed':
        return <ExclamationCircleOutlined style={{ color: '#ff4d4f' }} />;
      default:
        return <ClockCircleOutlined style={{ color: '#d9d9d9' }} />;
    }
  };

  if (loading) {
    return (
      <div style={{ padding: 24, textAlign: 'center' }}>
        <Spin size="large" />
        <div style={{ marginTop: 16 }}>加载中...</div>
      </div>
    );
  }

  if (!document || !processingStatus) {
    return (
      <div style={{ padding: 24 }}>
        <Alert
          message="文档不存在"
          description="请检查文档ID是否正确"
          type="error"
          showIcon
        />
      </div>
    );
  }

  return (
    <div style={{ padding: 24 }}>
      {/* 页面头部 */}
      <div style={{ marginBottom: 24 }}>
        <Space wrap>
          <Button 
            icon={<ArrowLeftOutlined />} 
            onClick={() => navigate(-1)}
          >
            <span className="desktop-only">返回</span>
          </Button>
          <Button 
            icon={<ReloadOutlined />} 
            onClick={refreshData}
            loading={refreshing}
          >
            <span className="desktop-only">刷新</span>
          </Button>
        </Space>
      </div>

      <Row gutter={[24, 24]}>
        {/* 左侧：文档信息 */}
        <Col xs={24} lg={12}>
          <Card title="文档信息" size="small">
            <Descriptions column={1} size="small" layout="vertical">
              <Descriptions.Item label="文件名">
                <div style={{ wordBreak: 'break-all' }}>
                  {document.original_name}
                </div>
              </Descriptions.Item>
              <Descriptions.Item label="文件大小">
                {formatFileSize(document.file_size)}
              </Descriptions.Item>
              <Descriptions.Item label="文件类型">
                {document.mime_type}
              </Descriptions.Item>
              <Descriptions.Item label="上传状态">
                <Tag color={document.status === 'completed' ? 'green' : 'orange'}>
                  {document.status === 'completed' ? '已完成' : '处理中'}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="描述">
                <div style={{ wordBreak: 'break-word' }}>
                  {document.description || '无描述'}
                </div>
              </Descriptions.Item>
              <Descriptions.Item label="上传时间">
                {new Date(document.created_at).toLocaleString()}
              </Descriptions.Item>
            </Descriptions>
          </Card>
        </Col>

        {/* 右侧：处理统计 */}
        <Col xs={24} lg={12}>
          <Card title="处理统计" size="small">
            <Row gutter={[16, 16]}>
              <Col xs={12} sm={12}>
                <Statistic
                  title="预处理进度"
                  value={processingStatus.progress}
                  precision={1}
                  suffix="%"
                  prefix={getStatusIcon(processingStatus.preprocessStatus)}
                  valueStyle={{ fontSize: '18px' }}
                />
              </Col>
              <Col xs={12} sm={12}>
                <Statistic
                  title="向量化进度"
                  value={processingStatus.vectorizationProgress}
                  precision={1}
                  suffix="%"
                  prefix={getStatusIcon(processingStatus.vectorizationStatus)}
                  valueStyle={{ fontSize: '18px' }}
                />
              </Col>
            </Row>
            <Divider />
            <Descriptions column={1} size="small" layout="vertical">
              <Descriptions.Item label="处理时间">
                {formatProcessingTime(processingStatus.processingTime)}
              </Descriptions.Item>
              <Descriptions.Item label="创建时间">
                {new Date(processingStatus.createdAt).toLocaleString()}
              </Descriptions.Item>
              <Descriptions.Item label="更新时间">
                {new Date(processingStatus.updatedAt).toLocaleString()}
              </Descriptions.Item>
              {processingStatus.completedAt && (
                <Descriptions.Item label="完成时间">
                  {new Date(processingStatus.completedAt).toLocaleString()}
                </Descriptions.Item>
              )}
            </Descriptions>
          </Card>
        </Col>
      </Row>

      {/* 处理状态详情 */}
      <Row gutter={[24, 24]} style={{ marginTop: 24 }}>
        <Col span={24}>
          <Card title="处理状态详情" size="small">
            <Row gutter={24}>
              {/* 预处理状态 */}
              <Col xs={24} lg={12}>
                <Card type="inner" title="预处理" size="small">
                  <Space direction="vertical" style={{ width: '100%' }}>
                    <div>
                      <Tag color={documentProcessingService.getStatusColor(processingStatus.preprocessStatus)}>
                        {documentProcessingService.getStatusText(processingStatus.preprocessStatus)}
                      </Tag>
                    </div>
                    
                    {processingStatus.progress > 0 && (
                      <Progress 
                        percent={Math.round(processingStatus.progress)} 
                        status={documentProcessingService.isFailed(processingStatus.preprocessStatus) ? 'exception' : 'active'}
                      />
                    )}
                    
                    {processingStatus.error && (
                      <Alert
                        message="处理错误"
                        description={processingStatus.error}
                        type="error"
                        showIcon
                      />
                    )}
                    
                    <Space wrap>
                      {processingStatus.preprocessStatus === 'not_started' && (
                        <Button 
                          type="primary" 
                          icon={<PlayCircleOutlined />}
                          onClick={handlePreprocess}
                          size="small"
                        >
                          <span className="desktop-only">开始预处理</span>
                        </Button>
                      )}
                      
                      {documentProcessingService.isFailed(processingStatus.preprocessStatus) && (
                        <Button 
                          icon={<ReloadOutlined />}
                          onClick={() => handleRetry('preprocess')}
                          size="small"
                        >
                          <span className="desktop-only">重试</span>
                        </Button>
                      )}
                      
                      {documentProcessingService.isProcessing(processingStatus.preprocessStatus) && (
                        <Button 
                          danger
                          icon={<StopOutlined />}
                          onClick={() => handleCancel('preprocess')}
                          size="small"
                        >
                          <span className="desktop-only">取消</span>
                        </Button>
                      )}
                    </Space>
                  </Space>
                </Card>
              </Col>

              {/* 向量化状态 */}
              <Col xs={24} lg={12}>
                <Card type="inner" title="向量化" size="small">
                  <Space direction="vertical" style={{ width: '100%' }}>
                    <div>
                      <Tag color={documentProcessingService.getStatusColor(processingStatus.vectorizationStatus)}>
                        {documentProcessingService.getStatusText(processingStatus.vectorizationStatus)}
                      </Tag>
                    </div>
                    
                    {processingStatus.vectorizationProgress > 0 && (
                      <Progress 
                        percent={Math.round(processingStatus.vectorizationProgress)} 
                        status={documentProcessingService.isFailed(processingStatus.vectorizationStatus) ? 'exception' : 'active'}
                      />
                    )}
                    
                    {processingStatus.vectorizationError && (
                      <Alert
                        message="向量化错误"
                        description={processingStatus.vectorizationError}
                        type="error"
                        showIcon
                      />
                    )}
                    
                    <Space wrap>
                      {processingStatus.vectorizationStatus === 'not_started' && (
                        <Button 
                          type="primary" 
                          icon={<ThunderboltOutlined />}
                          onClick={handleVectorize}
                          size="small"
                        >
                          <span className="desktop-only">开始向量化</span>
                        </Button>
                      )}
                      
                      {documentProcessingService.isFailed(processingStatus.vectorizationStatus) && (
                        <Button 
                          icon={<ReloadOutlined />}
                          onClick={() => handleRetry('vectorize')}
                          size="small"
                        >
                          <span className="desktop-only">重试</span>
                        </Button>
                      )}
                      
                      {documentProcessingService.isProcessing(processingStatus.vectorizationStatus) && (
                        <Button 
                          danger
                          icon={<StopOutlined />}
                          onClick={() => handleCancel('vectorize')}
                          size="small"
                        >
                          <span className="desktop-only">取消</span>
                        </Button>
                      )}
                    </Space>
                  </Space>
                </Card>
              </Col>
            </Row>
          </Card>
        </Col>
      </Row>

      {/* 处理时间线 */}
      <Row gutter={[24, 24]} style={{ marginTop: 24 }}>
        <Col span={24}>
          <Card title="处理历史" size="small">
            <Timeline>
              <Timeline.Item 
                dot={getStatusIcon('completed')}
                color="green"
              >
                <div>
                  <strong>文档上传完成</strong>
                  <div style={{ color: '#666', fontSize: '12px' }}>
                    {new Date(document.created_at).toLocaleString()}
                  </div>
                </div>
              </Timeline.Item>
              
              {processingStatus.preprocessStatus !== 'not_started' && (
                <Timeline.Item 
                  dot={getStatusIcon(processingStatus.preprocessStatus)}
                  color={documentProcessingService.getStatusColor(processingStatus.preprocessStatus)}
                >
                  <div>
                    <strong>预处理 - {documentProcessingService.getStatusText(processingStatus.preprocessStatus)}</strong>
                    <div style={{ color: '#666', fontSize: '12px' }}>
                      {new Date(processingStatus.createdAt).toLocaleString()}
                    </div>
                    {processingStatus.error && (
                      <div style={{ color: '#ff4d4f', fontSize: '12px' }}>
                        错误: {processingStatus.error}
                      </div>
                    )}
                  </div>
                </Timeline.Item>
              )}
              
              {processingStatus.vectorizationStatus !== 'not_started' && (
                <Timeline.Item 
                  dot={getStatusIcon(processingStatus.vectorizationStatus)}
                  color={documentProcessingService.getStatusColor(processingStatus.vectorizationStatus)}
                >
                  <div>
                    <strong>向量化 - {documentProcessingService.getStatusText(processingStatus.vectorizationStatus)}</strong>
                    <div style={{ color: '#666', fontSize: '12px' }}>
                      {new Date(processingStatus.updatedAt).toLocaleString()}
                    </div>
                    {processingStatus.vectorizationError && (
                      <div style={{ color: '#ff4d4f', fontSize: '12px' }}>
                        错误: {processingStatus.vectorizationError}
                      </div>
                    )}
                  </div>
                </Timeline.Item>
              )}
            </Timeline>
          </Card>
        </Col>
      </Row>

      {/* 响应式样式 */}
      <style dangerouslySetInnerHTML={{
        __html: `
          @media (max-width: 768px) {
            .desktop-only {
              display: none;
            }
            .ant-card-head-title {
              font-size: 14px !important;
            }
            .ant-descriptions-item-label {
              font-size: 12px !important;
            }
            .ant-descriptions-item-content {
              font-size: 12px !important;
            }
            .ant-statistic-title {
              font-size: 12px !important;
            }
            .ant-statistic-content {
              font-size: 16px !important;
            }
          }
          @media (max-width: 480px) {
            .ant-card {
              margin-bottom: 12px !important;
            }
            .ant-card-body {
              padding: 12px !important;
            }
            .ant-btn {
              padding: 4px 8px !important;
              font-size: 12px !important;
            }
            .ant-timeline-item-content {
              font-size: 12px !important;
            }
          }
        `
      }} />
    </div>
  );
};

export default DocumentProcessingDetail;