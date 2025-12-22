import React, { useState, useEffect } from 'react';
import {
  Card,
  Button,
  Progress,
  Tag,
  Space,
  Alert,
  Descriptions,
  Modal,
  Select,
  InputNumber,
  message,
  Tooltip,
  Divider
} from 'antd';
import {
  PlayCircleOutlined,
  PauseCircleOutlined,
  ReloadOutlined,
  StopOutlined,
  FileTextOutlined,
  ClockCircleOutlined,
  CheckCircleOutlined,
  ExclamationCircleOutlined
} from '@ant-design/icons';
import { useProcessingStatus, useProcessingActions, ProcessingUtils } from '../hooks/useProcessing';
import type { ProcessingStatusResponse, ProcessDocumentAsyncRequest } from '../types/processing';

interface DocumentProcessingProps {
  documentId: string;
  documentName?: string;
  onStatusChange?: (status: ProcessingStatusResponse) => void;
  showActions?: boolean;
  autoRefresh?: boolean;
}

const DocumentProcessing: React.FC<DocumentProcessingProps> = ({
  documentId,
  documentName,
  onStatusChange,
  showActions = true,
  autoRefresh = true
}) => {
  const [asyncModalVisible, setAsyncModalVisible] = useState(false);
  const [asyncOptions, setAsyncOptions] = useState<ProcessDocumentAsyncRequest>({ priority: 1 });

  const {
    status,
    loading: statusLoading,
    error: statusError,
    refresh,
    startPolling,
    stopPolling
  } = useProcessingStatus(documentId, autoRefresh);

  const {
    processDocument,
    processDocumentAsync,
    reprocessDocument,
    loading: actionLoading,
    error: actionError
  } = useProcessingActions();

  // 通知父组件状态变化
  useEffect(() => {
    if (status && onStatusChange) {
      onStatusChange(status);
    }
  }, [status, onStatusChange]);

  // 处理同步处理
  const handleProcess = async () => {
    try {
      await processDocument(documentId);
      startPolling({ stopOnComplete: true });
    } catch (error) {
      console.error('Process failed:', error);
    }
  };

  // 处理异步处理
  const handleAsyncProcess = async () => {
    try {
      const taskId = await processDocumentAsync(documentId, asyncOptions);
      message.success(`异步任务已创建: ${taskId}`);
      setAsyncModalVisible(false);
      startPolling({ stopOnComplete: true });
    } catch (error) {
      console.error('Async process failed:', error);
    }
  };

  // 处理重新处理
  const handleReprocess = async () => {
    Modal.confirm({
      title: '确认重新处理',
      content: '重新处理将覆盖现有的处理结果，是否继续？',
      onOk: async () => {
        try {
          await reprocessDocument(documentId);
          startPolling({ stopOnComplete: true });
        } catch (error) {
          console.error('Reprocess failed:', error);
        }
      }
    });
  };

  // 渲染状态标签
  const renderStatusTag = () => {
    if (!status) return null;

    const color = ProcessingUtils.getStatusColor(status.status);
    const text = ProcessingUtils.getStatusText(status.status);
    
    return <Tag color={color}>{text}</Tag>;
  };

  // 渲染进度条
  const renderProgress = () => {
    if (!status) return null;

    const { progress, status: currentStatus } = status;
    const isProcessing = ProcessingUtils.isProcessing(currentStatus);
    const isFailed = ProcessingUtils.isFailed(currentStatus);
    
    return (
      <Progress
        percent={Math.round(progress)}
        status={isFailed ? 'exception' : isProcessing ? 'active' : 'success'}
        showInfo={true}
        format={(percent) => `${percent}%`}
      />
    );
  };

  // 渲染操作按钮
  const renderActions = () => {
    if (!showActions) return null;

    const isProcessing = status && ProcessingUtils.isProcessing(status.status);
    const isCompleted = status && ProcessingUtils.isCompleted(status.status);
    const isFailed = status && ProcessingUtils.isFailed(status.status);

    return (
      <Space>
        <Button
          type="primary"
          icon={<PlayCircleOutlined />}
          onClick={handleProcess}
          loading={actionLoading}
          disabled={isProcessing}
        >
          开始处理
        </Button>
        
        <Button
          icon={<PlayCircleOutlined />}
          onClick={() => setAsyncModalVisible(true)}
          loading={actionLoading}
          disabled={isProcessing}
        >
          异步处理
        </Button>

        {(isCompleted || isFailed) && (
          <Button
            icon={<ReloadOutlined />}
            onClick={handleReprocess}
            loading={actionLoading}
          >
            重新处理
          </Button>
        )}

        {isProcessing && (
          <Button
            icon={<StopOutlined />}
            onClick={stopPolling}
            danger
          >
            停止监控
          </Button>
        )}

        <Button
          icon={<ReloadOutlined />}
          onClick={refresh}
          loading={statusLoading}
        >
          刷新状态
        </Button>
      </Space>
    );
  };

  // 渲染详细信息
  const renderDetails = () => {
    if (!status) return null;

    const items = [
      {
        key: 'document_id',
        label: '文档ID',
        children: status.document_id
      },
      {
        key: 'status',
        label: '处理状态',
        children: renderStatusTag()
      },
      {
        key: 'progress',
        label: '处理进度',
        children: `${Math.round(status.progress)}%`
      }
    ];

    if (status.processed_size > 0 || status.total_size > 0) {
      items.push({
        key: 'size',
        label: '处理大小',
        children: `${ProcessingUtils.formatFileSize(status.processed_size)} / ${ProcessingUtils.formatFileSize(status.total_size)}`
      });
    }

    if (status.started_at) {
      items.push({
        key: 'started_at',
        label: '开始时间',
        children: new Date(status.started_at).toLocaleString()
      });
    }

    if (status.completed_at) {
      items.push({
        key: 'completed_at',
        label: '完成时间',
        children: new Date(status.completed_at).toLocaleString()
      });
    }

    if (status.error_message) {
      items.push({
        key: 'error',
        label: '错误信息',
        children: (
          <Alert
            message={status.error_message}
            type="error"
            showIcon
            size="small"
          />
        )
      });
    }

    return (
      <Descriptions
        size="small"
        column={1}
        items={items}
      />
    );
  };

  return (
    <>
      <Card
        title={
          <Space>
            <FileTextOutlined />
            {documentName || `文档 ${documentId}`}
            {renderStatusTag()}
          </Space>
        }
        extra={
          <Space>
            {status && ProcessingUtils.isProcessing(status.status) && (
              <Tooltip title="处理中">
                <ClockCircleOutlined spin style={{ color: '#1890ff' }} />
              </Tooltip>
            )}
            {status && ProcessingUtils.isCompleted(status.status) && (
              <Tooltip title="处理完成">
                <CheckCircleOutlined style={{ color: '#52c41a' }} />
              </Tooltip>
            )}
            {status && ProcessingUtils.isFailed(status.status) && (
              <Tooltip title="处理失败">
                <ExclamationCircleOutlined style={{ color: '#ff4d4f' }} />
              </Tooltip>
            )}
          </Space>
        }
        loading={statusLoading}
      >
        {statusError && (
          <Alert
            message="获取状态失败"
            description={statusError}
            type="error"
            showIcon
            style={{ marginBottom: 16 }}
          />
        )}

        {actionError && (
          <Alert
            message="操作失败"
            description={actionError}
            type="error"
            showIcon
            style={{ marginBottom: 16 }}
          />
        )}

        {status && (
          <>
            {renderProgress()}
            <Divider />
            {renderDetails()}
          </>
        )}

        {renderActions()}
      </Card>

      {/* 异步处理配置模态框 */}
      <Modal
        title="异步处理配置"
        open={asyncModalVisible}
        onOk={handleAsyncProcess}
        onCancel={() => setAsyncModalVisible(false)}
        confirmLoading={actionLoading}
      >
        <Space direction="vertical" style={{ width: '100%' }}>
          <div>
            <label>任务优先级 (1-10，数字越大优先级越高):</label>
            <InputNumber
              min={1}
              max={10}
              value={asyncOptions.priority}
              onChange={(value) => setAsyncOptions({ ...asyncOptions, priority: value || 1 })}
              style={{ width: '100%', marginTop: 8 }}
            />
          </div>
        </Space>
      </Modal>
    </>
  );
};

export default DocumentProcessing;