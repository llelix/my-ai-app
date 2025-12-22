import React, { useState, useEffect } from 'react';
import {
  Card,
  Button,
  Table,
  Space,
  Select,
  InputNumber,
  Checkbox,
  Progress,
  Tag,
  Alert,
  Modal,
  message,
  Tooltip
} from 'antd';
import {
  PlayCircleOutlined,
  ReloadOutlined,
  DeleteOutlined,
  SettingOutlined,
  FileTextOutlined
} from '@ant-design/icons';
import { useProcessingActions, ProcessingUtils } from '../hooks/useProcessing';
import type { Document } from '../types';
import type { ProcessingStatusResponse } from '../types/processing';

interface BatchProcessingProps {
  documents: Document[];
  onDocumentStatusChange?: (documentId: string, status: ProcessingStatusResponse) => void;
}

interface DocumentWithStatus extends Document {
  processing_status?: ProcessingStatusResponse;
  selected?: boolean;
}

const BatchProcessing: React.FC<BatchProcessingProps> = ({
  documents,
  onDocumentStatusChange
}) => {
  const [documentList, setDocumentList] = useState<DocumentWithStatus[]>([]);
  const [selectedDocuments, setSelectedDocuments] = useState<string[]>([]);
  const [batchOptions, setBatchOptions] = useState({
    async: true,
    priority: 1
  });
  const [optionsModalVisible, setOptionsModalVisible] = useState(false);

  const {
    batchProcess,
    loading: batchLoading,
    error: batchError
  } = useProcessingActions();

  // 初始化文档列表
  useEffect(() => {
    setDocumentList(documents.map(doc => ({ ...doc, selected: false })));
  }, [documents]);

  // 处理文档选择
  const handleDocumentSelect = (documentId: string, selected: boolean) => {
    setDocumentList(prev => 
      prev.map(doc => 
        doc.id.toString() === documentId ? { ...doc, selected } : doc
      )
    );

    if (selected) {
      setSelectedDocuments(prev => [...prev, documentId]);
    } else {
      setSelectedDocuments(prev => prev.filter(id => id !== documentId));
    }
  };

  // 处理全选
  const handleSelectAll = (checked: boolean) => {
    const eligibleDocs = documentList.filter(doc => 
      doc.status === 'completed' && 
      (!doc.processing_status || !ProcessingUtils.isProcessing(doc.processing_status.status))
    );

    if (checked) {
      const newSelected = eligibleDocs.map(doc => doc.id.toString());
      setSelectedDocuments(newSelected);
      setDocumentList(prev => 
        prev.map(doc => ({
          ...doc,
          selected: eligibleDocs.some(eligible => eligible.id === doc.id)
        }))
      );
    } else {
      setSelectedDocuments([]);
      setDocumentList(prev => prev.map(doc => ({ ...doc, selected: false })));
    }
  };

  // 开始批量处理
  const handleBatchProcess = async () => {
    if (selectedDocuments.length === 0) {
      message.warning('请选择要处理的文档');
      return;
    }

    try {
      await batchProcess(selectedDocuments, batchOptions.async, batchOptions.priority);
      
      // 清空选择
      setSelectedDocuments([]);
      setDocumentList(prev => prev.map(doc => ({ ...doc, selected: false })));
    } catch (error) {
      console.error('Batch process failed:', error);
    }
  };

  // 移除选中的文档
  const handleRemoveSelected = () => {
    setSelectedDocuments([]);
    setDocumentList(prev => prev.map(doc => ({ ...doc, selected: false })));
  };

  // 渲染状态标签
  const renderStatusTag = (status?: ProcessingStatusResponse) => {
    if (!status) return <Tag color="default">未处理</Tag>;

    const color = ProcessingUtils.getStatusColor(status.status);
    const text = ProcessingUtils.getStatusText(status.status);
    
    return <Tag color={color}>{text}</Tag>;
  };

  // 渲染进度
  const renderProgress = (status?: ProcessingStatusResponse) => {
    if (!status || status.progress === 0) return null;

    const isProcessing = ProcessingUtils.isProcessing(status.status);
    const isFailed = ProcessingUtils.isFailed(status.status);
    
    return (
      <Progress
        percent={Math.round(status.progress)}
        size="small"
        status={isFailed ? 'exception' : isProcessing ? 'active' : 'success'}
        showInfo={false}
        style={{ width: 100 }}
      />
    );
  };

  // 表格列定义
  const columns = [
    {
      title: (
        <Checkbox
          checked={selectedDocuments.length > 0}
          indeterminate={selectedDocuments.length > 0 && selectedDocuments.length < documentList.length}
          onChange={(e) => handleSelectAll(e.target.checked)}
        >
          全选
        </Checkbox>
      ),
      dataIndex: 'selected',
      width: 80,
      render: (_: any, record: DocumentWithStatus) => {
        const isEligible = record.status === 'completed' && 
          (!record.processing_status || !ProcessingUtils.isProcessing(record.processing_status.status));
        
        return (
          <Checkbox
            checked={record.selected}
            disabled={!isEligible}
            onChange={(e) => handleDocumentSelect(record.id.toString(), e.target.checked)}
          />
        );
      }
    },
    {
      title: '文档名称',
      dataIndex: 'name',
      render: (text: string, record: DocumentWithStatus) => (
        <Space>
          <FileTextOutlined />
          <Tooltip title={record.original_name}>
            {text}
          </Tooltip>
        </Space>
      )
    },
    {
      title: '文件大小',
      dataIndex: 'file_size',
      width: 120,
      render: (size: number) => ProcessingUtils.formatFileSize(size)
    },
    {
      title: '上传状态',
      dataIndex: 'status',
      width: 100,
      render: (status: string) => {
        const colorMap: Record<string, string> = {
          'uploading': 'processing',
          'processing': 'warning',
          'completed': 'success',
          'failed': 'error'
        };
        return <Tag color={colorMap[status] || 'default'}>{status}</Tag>;
      }
    },
    {
      title: '处理状态',
      dataIndex: 'processing_status',
      width: 120,
      render: (status: ProcessingStatusResponse) => renderStatusTag(status)
    },
    {
      title: '处理进度',
      dataIndex: 'processing_status',
      width: 120,
      render: (status: ProcessingStatusResponse) => renderProgress(status)
    },
    {
      title: '上传时间',
      dataIndex: 'created_at',
      width: 180,
      render: (time: string) => new Date(time).toLocaleString()
    }
  ];

  const selectedCount = selectedDocuments.length;
  const totalCount = documentList.length;
  const eligibleCount = documentList.filter(doc => 
    doc.status === 'completed' && 
    (!doc.processing_status || !ProcessingUtils.isProcessing(doc.processing_status.status))
  ).length;

  return (
    <>
      <Card
        title="批量文档处理"
        extra={
          <Space>
            <Button
              icon={<SettingOutlined />}
              onClick={() => setOptionsModalVisible(true)}
            >
              处理选项
            </Button>
            <Button
              type="primary"
              icon={<PlayCircleOutlined />}
              onClick={handleBatchProcess}
              loading={batchLoading}
              disabled={selectedCount === 0}
            >
              开始处理 ({selectedCount})
            </Button>
          </Space>
        }
      >
        {batchError && (
          <Alert
            message="批量处理失败"
            description={batchError}
            type="error"
            showIcon
            style={{ marginBottom: 16 }}
          />
        )}

        <div style={{ marginBottom: 16 }}>
          <Space>
            <span>已选择: {selectedCount} / {totalCount}</span>
            <span>可处理: {eligibleCount}</span>
            {selectedCount > 0 && (
              <Button
                size="small"
                icon={<DeleteOutlined />}
                onClick={handleRemoveSelected}
              >
                清空选择
              </Button>
            )}
          </Space>
        </div>

        <Table
          columns={columns}
          dataSource={documentList}
          rowKey="id"
          size="small"
          pagination={{
            pageSize: 10,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total, range) => `第 ${range[0]}-${range[1]} 条，共 ${total} 条`
          }}
        />
      </Card>

      {/* 处理选项模态框 */}
      <Modal
        title="批量处理选项"
        open={optionsModalVisible}
        onOk={() => setOptionsModalVisible(false)}
        onCancel={() => setOptionsModalVisible(false)}
      >
        <Space direction="vertical" style={{ width: '100%' }}>
          <div>
            <Checkbox
              checked={batchOptions.async}
              onChange={(e) => setBatchOptions({ ...batchOptions, async: e.target.checked })}
            >
              异步处理
            </Checkbox>
            <div style={{ color: '#666', fontSize: '12px', marginTop: 4 }}>
              异步处理可以避免长时间等待，适合大量文档处理
            </div>
          </div>

          <div>
            <label>任务优先级 (1-10):</label>
            <InputNumber
              min={1}
              max={10}
              value={batchOptions.priority}
              onChange={(value) => setBatchOptions({ ...batchOptions, priority: value || 1 })}
              style={{ width: '100%', marginTop: 8 }}
            />
            <div style={{ color: '#666', fontSize: '12px', marginTop: 4 }}>
              数字越大优先级越高，高优先级任务会优先处理
            </div>
          </div>
        </Space>
      </Modal>
    </>
  );
};

export default BatchProcessing;