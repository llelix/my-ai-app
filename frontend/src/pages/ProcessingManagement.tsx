import React, { useState, useEffect } from 'react';
import { Tabs, Card, Space, Button, message } from 'antd';
import {
  DashboardOutlined,
  FileTextOutlined,
  AppstoreOutlined,
  SettingOutlined
} from '@ant-design/icons';
import {
  ProcessingDashboard,
  DocumentProcessing,
  BatchProcessing
} from '../components';
import { documentService } from '../services';
import type { Document } from '../types';
import type { ProcessingStatusResponse } from '../types/processing';

const { TabPane } = Tabs;

const ProcessingManagement: React.FC = () => {
  const [documents, setDocuments] = useState<Document[]>([]);
  const [loading, setLoading] = useState(false);
  const [selectedDocumentId, setSelectedDocumentId] = useState<string | null>(null);

  // 加载文档列表
  const loadDocuments = async () => {
    setLoading(true);
    try {
      const response = await documentService.getDocuments();
      if (response.data) {
        setDocuments(response.data);
      }
    } catch (error) {
      message.error('加载文档列表失败');
      console.error('Error loading documents:', error);
    } finally {
      setLoading(false);
    }
  };

  // 初始加载
  useEffect(() => {
    loadDocuments();
  }, []);

  // 处理文档状态变化
  const handleDocumentStatusChange = (documentId: string, status: ProcessingStatusResponse) => {
    console.log(`Document ${documentId} status changed:`, status);
    // 这里可以更新文档列表中的状态
    // 或者触发其他相关操作
  };

  // 渲染单个文档处理页面
  const renderSingleDocumentProcessing = () => {
    return (
      <Card title="单个文档处理">
        <Space direction="vertical" style={{ width: '100%' }}>
          <div>
            <label>选择文档:</label>
            <select
              value={selectedDocumentId || ''}
              onChange={(e) => setSelectedDocumentId(e.target.value || null)}
              style={{ width: '100%', padding: '8px', marginTop: '8px' }}
            >
              <option value="">请选择文档</option>
              {documents.map(doc => (
                <option key={doc.id} value={doc.id.toString()}>
                  {doc.name} ({doc.original_name})
                </option>
              ))}
            </select>
          </div>

          {selectedDocumentId && (
            <DocumentProcessing
              documentId={selectedDocumentId}
              documentName={documents.find(d => d.id.toString() === selectedDocumentId)?.name}
              onStatusChange={handleDocumentStatusChange}
              showActions={true}
              autoRefresh={true}
            />
          )}
        </Space>
      </Card>
    );
  };

  // 渲染批量处理页面
  const renderBatchProcessing = () => {
    return (
      <BatchProcessing
        documents={documents}
        onDocumentStatusChange={handleDocumentStatusChange}
      />
    );
  };

  // 渲染仪表板页面
  const renderDashboard = () => {
    return <ProcessingDashboard />;
  };

  return (
    <div style={{ padding: '24px' }}>
      <div style={{ marginBottom: '24px', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <h1>文档处理管理</h1>
        <Button
          onClick={loadDocuments}
          loading={loading}
        >
          刷新文档列表
        </Button>
      </div>

      <Tabs defaultActiveKey="dashboard" size="large">
        <TabPane
          tab={
            <Space>
              <DashboardOutlined />
              仪表板
            </Space>
          }
          key="dashboard"
        >
          {renderDashboard()}
        </TabPane>

        <TabPane
          tab={
            <Space>
              <FileTextOutlined />
              单个处理
            </Space>
          }
          key="single"
        >
          {renderSingleDocumentProcessing()}
        </TabPane>

        <TabPane
          tab={
            <Space>
              <AppstoreOutlined />
              批量处理
            </Space>
          }
          key="batch"
        >
          {renderBatchProcessing()}
        </TabPane>
      </Tabs>
    </div>
  );
};

export default ProcessingManagement;