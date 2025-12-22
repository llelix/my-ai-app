import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { 
  Upload, 
  Table, 
  Button, 
  message, 
  Modal, 
  Input, 
  Space,
  Popconfirm,
  Progress,
  Switch,
  Tooltip
} from 'antd';
import { 
  UploadOutlined, 
  DownloadOutlined, 
  DeleteOutlined, 
  EditOutlined,
  SettingOutlined,
  EyeOutlined
} from '@ant-design/icons';
import { documentService, documentProcessingService } from '../services';
import { DocumentProcessingButtons, DocumentProcessingStatusIndicator } from '../components';
import type { Document, DocumentProcessingStatus } from '../types';

const DocumentManagement: React.FC = () => {
  const navigate = useNavigate();
  const [documents, setDocuments] = useState<Document[]>([]);
  const [loading, setLoading] = useState(false);
  const [editModal, setEditModal] = useState<{ visible: boolean; document?: Document }>({ visible: false });
  const [description, setDescription] = useState('');
  const [uploadProgress, setUploadProgress] = useState<{ [key: string]: number }>({});
  const [processingStatuses, setProcessingStatuses] = useState<{ [key: string]: DocumentProcessingStatus }>({});
  const [showProcessingColumns, setShowProcessingColumns] = useState(true);
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([]);

  useEffect(() => {
    loadDocuments();
    
    // 组件卸载时清理轮询
    return () => {
      documentProcessingService.cleanup();
    };
  }, []);

  const loadDocuments = async () => {
    setLoading(true);
    try {
      const response = await documentService.getDocuments();
      // 根据实际API返回结构调整
      const documentsData = response.data?.items || response.data || [];
      const docs = Array.isArray(documentsData) ? documentsData : [];
      setDocuments(docs);
      
      // 加载每个文档的处理状态
      await loadProcessingStatuses(docs);
    } catch (error) {
      console.error('加载文档失败:', error);
      message.error('加载文档失败');
    } finally {
      setLoading(false);
    }
  };

  // 加载文档处理状态
  const loadProcessingStatuses = async (docs: Document[]) => {
    const statusPromises = docs.map(async (doc) => {
      try {
        const status = await documentProcessingService.getProcessingStatus(doc.id.toString());
        return { docId: doc.id.toString(), status };
      } catch (error) {
        // 如果获取状态失败，返回默认状态
        return {
          docId: doc.id.toString(),
          status: {
            documentId: doc.id.toString(),
            preprocessStatus: 'not_started' as const,
            vectorizationStatus: 'not_started' as const,
            progress: 0,
            vectorizationProgress: 0,
            createdAt: new Date().toISOString(),
            updatedAt: new Date().toISOString()
          }
        };
      }
    });

    const results = await Promise.all(statusPromises);
    const statusMap: { [key: string]: DocumentProcessingStatus } = {};
    results.forEach(({ docId, status }) => {
      statusMap[docId] = status;
    });
    setProcessingStatuses(statusMap);
  };

  // 处理状态更新
  const handleStatusUpdate = (documentId: string, status: DocumentProcessingStatus) => {
    setProcessingStatuses(prev => ({
      ...prev,
      [documentId]: status
    }));
  };

  // 批量处理操作
  const handleBatchPreprocess = async () => {
    if (selectedRowKeys.length === 0) {
      message.warning('请选择要处理的文档');
      return;
    }

    try {
      const documentIds = selectedRowKeys.map(key => key.toString());
      await documentProcessingService.batchTriggerPreprocessing(documentIds);
      message.success(`已开始批量预处理 ${documentIds.length} 个文档`);
      
      // 重新加载状态
      await loadProcessingStatuses(documents);
    } catch (error) {
      console.error('批量预处理失败:', error);
      message.error('批量预处理失败');
    }
  };

  const handleUpload = async (file: File) => {
    const fileKey = `${file.name}_${file.size}`;
    const fileSizeMB = file.size / (1024 * 1024); // 转换为MB
    
    try {
      if (fileSizeMB > 50) {
        // 文件大于50MB，使用分片上传（支持断点续传和秒传）
        await documentService.uploadWithResume(file, (progress) => {
          setUploadProgress(prev => ({ ...prev, [fileKey]: progress }));
        });
        
        if (uploadProgress[fileKey] === 100) {
          message.success('上传成功（秒传）');
        } else {
          message.success('上传成功');
        }
      } else {
        // 文件小于等于50MB，使用传统上传
        await documentService.upload(file);
        message.success('上传成功');
      }
      
      setUploadProgress(prev => {
        const newProgress = { ...prev };
        delete newProgress[fileKey];
        return newProgress;
      });
      
      loadDocuments();
    } catch (error: any) {
      message.error(`上传失败: ${error.message || '未知错误'}`);
      setUploadProgress(prev => {
        const newProgress = { ...prev };
        delete newProgress[fileKey];
        return newProgress;
      });
    }
    
    return false;
  };

  const handleDelete = async (id: number) => {
    try {
      await documentService.deleteDocument(id);
      message.success('删除成功');
      loadDocuments();
    } catch (error) {
      message.error('删除失败');
    }
  };

  const handleEditDescription = (doc: Document) => {
    setEditModal({ visible: true, document: doc });
    setDescription(doc.description || '');
  };

  const handleSaveDescription = async () => {
    if (!editModal.document) return;
    
    try {
      await documentService.updateDescription(editModal.document.id, description);
      message.success('更新成功');
      setEditModal({ visible: false });
      loadDocuments();
    } catch (error) {
      message.error('更新失败');
    }
  };

  const formatFileSize = (bytes: number) => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  const getStatusBadge = (status: string) => {
    const statusMap: { [key: string]: { color: string; text: string } } = {
      completed: { color: 'green', text: '已完成' },
      uploading: { color: 'blue', text: '上传中' },
      failed: { color: 'red', text: '失败' }
    };
    const statusInfo = statusMap[status] || { color: 'gray', text: '未知' };
    return <span style={{ color: statusInfo.color }}>{statusInfo.text}</span>;
  };

  const columns = [
    {
      title: '文件名',
      dataIndex: 'original_name',
      key: 'original_name',
      width: 200,
    },
    {
      title: '大小',
      dataIndex: 'file_size',
      key: 'file_size',
      width: 100,
      render: (size: number) => formatFileSize(size),
    },
    {
      title: '类型',
      dataIndex: 'mime_type',
      key: 'mime_type',
      width: 120,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 80,
      render: (status: string) => getStatusBadge(status),
    },
    ...(showProcessingColumns ? [
      {
        title: '处理状态',
        key: 'processing_status',
        width: 200,
        render: (_: any, record: Document) => {
          const status = processingStatuses[record.id.toString()];
          return status ? (
            <DocumentProcessingStatusIndicator 
              status={status}
              showProgress={true}
              showBothStatuses={true}
              size="small"
            />
          ) : (
            <span style={{ color: '#999' }}>加载中...</span>
          );
        },
      },
      {
        title: '处理操作',
        key: 'processing_actions',
        width: 300,
        render: (_: any, record: Document) => {
          const status = processingStatuses[record.id.toString()];
          return status ? (
            <DocumentProcessingButtons
              documentId={record.id.toString()}
              initialStatus={status}
              onStatusChange={(newStatus) => handleStatusUpdate(record.id.toString(), newStatus)}
              size="small"
              showProgress={false}
              showRetry={true}
            />
          ) : null;
        },
      }
    ] : []),
    {
      title: '描述',
      dataIndex: 'description',
      key: 'description',
      width: 150,
      render: (text: string) => text || '-',
    },
    {
      title: '上传时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 150,
      render: (date: string) => new Date(date).toLocaleString(),
    },
    {
      title: '操作',
      key: 'actions',
      width: 250,
      render: (_: any, record: Document) => (
        <Space>
          <Button
            size="small"
            icon={<EyeOutlined />}
            onClick={() => navigate(`/documents/${record.id}/processing`)}
          >
            详情
          </Button>
          <Button
            size="small"
            icon={<DownloadOutlined />}
            onClick={() => window.open(documentService.getDownloadUrl(record.id))}
          >
            下载
          </Button>
          <Button
            size="small"
            icon={<EditOutlined />}
            onClick={() => handleEditDescription(record)}
          >
            编辑
          </Button>
          <Popconfirm
            title="确定删除此文档吗？"
            onConfirm={() => handleDelete(record.id)}
          >
            <Button size="small" danger icon={<DeleteOutlined />}>
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div style={{ padding: 24 }}>
      <div style={{ marginBottom: 16 }}>
        <div style={{ 
          display: 'flex', 
          justifyContent: 'space-between', 
          alignItems: 'center',
          flexWrap: 'wrap',
          gap: '12px'
        }}>
          <Space wrap>
            <Upload
              beforeUpload={handleUpload}
              showUploadList={false}
              multiple
            >
              <Button icon={<UploadOutlined />}>上传文档</Button>
            </Upload>
            
            {selectedRowKeys.length > 0 && (
              <Button 
                type="primary" 
                onClick={handleBatchPreprocess}
                disabled={selectedRowKeys.length === 0}
              >
                批量预处理 ({selectedRowKeys.length})
              </Button>
            )}
          </Space>

          <Space wrap>
            <Tooltip title="显示/隐藏处理状态列">
              <Switch
                checked={showProcessingColumns}
                onChange={setShowProcessingColumns}
                checkedChildren="处理列"
                unCheckedChildren="处理列"
                size="small"
              />
            </Tooltip>
            
            <Tooltip title="刷新文档列表">
              <Button 
                icon={<SettingOutlined />} 
                onClick={loadDocuments}
                loading={loading}
                size="small"
              >
                <span className="desktop-only">刷新</span>
              </Button>
            </Tooltip>
          </Space>
        </div>
      </div>

      {/* 上传进度显示 */}
      {Object.keys(uploadProgress).length > 0 && (
        <div style={{ marginBottom: 16 }}>
          {Object.entries(uploadProgress).map(([fileKey, progress]) => (
            <div key={fileKey} style={{ marginBottom: 8 }}>
              <div style={{ marginBottom: 4 }}>{fileKey.split('_')[0]}</div>
              <Progress percent={Math.round(progress)} status="active" />
            </div>
          ))}
        </div>
      )}

      <Table
        columns={columns}
        dataSource={documents}
        rowKey="id"
        loading={loading}
        scroll={{ x: showProcessingColumns ? 1400 : 800 }}
        rowSelection={showProcessingColumns ? {
          selectedRowKeys,
          onChange: setSelectedRowKeys,
          getCheckboxProps: (record: Document) => ({
            disabled: record.status !== 'completed', // 只允许选择已完成上传的文档
          }),
        } : undefined}
        pagination={{
          showSizeChanger: true,
          showQuickJumper: true,
          showTotal: (total, range) => `第 ${range[0]}-${range[1]} 条，共 ${total} 条`,
          responsive: true,
        }}
        size="small"
      />

      <Modal
        title="编辑描述"
        open={editModal.visible}
        onOk={handleSaveDescription}
        onCancel={() => setEditModal({ visible: false })}
      >
        <Input.TextArea
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          placeholder="请输入文档描述"
          rows={4}
        />
      </Modal>

      {/* 响应式样式 */}
      <style dangerouslySetInnerHTML={{
        __html: `
          @media (max-width: 768px) {
            .desktop-only {
              display: none;
            }
            .ant-table-thead > tr > th {
              padding: 8px 4px !important;
              font-size: 12px !important;
            }
            .ant-table-tbody > tr > td {
              padding: 8px 4px !important;
              font-size: 12px !important;
            }
          }
          @media (max-width: 480px) {
            .ant-table-thead > tr > th {
              padding: 6px 2px !important;
              font-size: 11px !important;
            }
            .ant-table-tbody > tr > td {
              padding: 6px 2px !important;
              font-size: 11px !important;
            }
            .ant-btn {
              padding: 4px 8px !important;
              font-size: 12px !important;
            }
          }
        `
      }} />
    </div>
  );
};

export default DocumentManagement;
