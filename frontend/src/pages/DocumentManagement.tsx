import React, { useState, useEffect } from 'react';
import { 
  Upload, 
  Table, 
  Button, 
  message, 
  Modal, 
  Input, 
  Space,
  Popconfirm,
  Progress
} from 'antd';
import { 
  UploadOutlined, 
  DownloadOutlined, 
  DeleteOutlined, 
  EditOutlined 
} from '@ant-design/icons';
import { documentService } from '../services/documentService';
import type { Document } from '../types';

const DocumentManagement: React.FC = () => {
  const [documents, setDocuments] = useState<Document[]>([]);
  const [loading, setLoading] = useState(false);
  const [editModal, setEditModal] = useState<{ visible: boolean; document?: Document }>({ visible: false });
  const [description, setDescription] = useState('');
  const [uploadProgress, setUploadProgress] = useState<{ [key: string]: number }>({});

  useEffect(() => {
    loadDocuments();
  }, []);

  const loadDocuments = async () => {
    setLoading(true);
    try {
      const response = await documentService.getDocuments();
      // 根据实际API返回结构调整
      const documentsData = response.data?.items || response.data || [];
      setDocuments(Array.isArray(documentsData) ? documentsData : []);
    } catch (error) {
      console.error('加载文档失败:', error);
      message.error('加载文档失败');
    } finally {
      setLoading(false);
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
    },
    {
      title: '大小',
      dataIndex: 'file_size',
      key: 'file_size',
      render: (size: number) => formatFileSize(size),
    },
    {
      title: '类型',
      dataIndex: 'mime_type',
      key: 'mime_type',
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => getStatusBadge(status),
    },
    {
      title: '描述',
      dataIndex: 'description',
      key: 'description',
      render: (text: string) => text || '-',
    },
    {
      title: '上传时间',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (date: string) => new Date(date).toLocaleString(),
    },
    {
      title: '操作',
      key: 'actions',
      render: (_: any, record: Document) => (
        <Space>
          <Button
            icon={<DownloadOutlined />}
            onClick={() => window.open(documentService.getDownloadUrl(record.id))}
          >
            下载
          </Button>
          <Button
            icon={<EditOutlined />}
            onClick={() => handleEditDescription(record)}
          >
            编辑
          </Button>
          <Popconfirm
            title="确定删除此文档吗？"
            onConfirm={() => handleDelete(record.id)}
          >
            <Button danger icon={<DeleteOutlined />}>
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
        <Upload
          beforeUpload={handleUpload}
          showUploadList={false}
          multiple
        >
          <Button icon={<UploadOutlined />}>上传文档</Button>
        </Upload>
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
    </div>
  );
};

export default DocumentManagement;
