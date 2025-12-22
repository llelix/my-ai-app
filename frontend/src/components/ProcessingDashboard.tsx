import React, { useState, useEffect } from 'react';
import {
  Card,
  Row,
  Col,
  Statistic,
  Progress,
  Table,
  Tag,
  Space,
  Button,
  Alert,
  Descriptions,
  Divider
} from 'antd';
import {
  ReloadOutlined,
  ClockCircleOutlined,
  CheckCircleOutlined,
  ExclamationCircleOutlined,
  FileTextOutlined,
  ThunderboltOutlined,
  TeamOutlined
} from '@ant-design/icons';
import { processingService } from '../services';
import { ProcessingUtils } from '../hooks/useProcessing';
import type {
  QueueStatsResponse,
  ProcessingStatisticsResponse,
  SupportedFormatsResponse
} from '../types/processing';

const ProcessingDashboard: React.FC = () => {
  const [queueStats, setQueueStats] = useState<QueueStatsResponse | null>(null);
  const [processingStats, setProcessingStats] = useState<ProcessingStatisticsResponse | null>(null);
  const [supportedFormats, setSupportedFormats] = useState<SupportedFormatsResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // 加载统计数据
  const loadStats = async () => {
    setLoading(true);
    setError(null);

    try {
      const [queueData, statsData, formatsData] = await Promise.all([
        processingService.getQueueStats(),
        processingService.getProcessingStatistics(),
        processingService.getSupportedFormats()
      ]);

      setQueueStats(queueData);
      setProcessingStats(statsData);
      setSupportedFormats(formatsData);
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to load statistics';
      setError(errorMessage);
      console.error('Error loading processing stats:', err);
    } finally {
      setLoading(false);
    }
  };

  // 初始加载
  useEffect(() => {
    loadStats();
    
    // 设置定时刷新
    const interval = setInterval(loadStats, 30000); // 30秒刷新一次
    
    return () => clearInterval(interval);
  }, []);

  // 渲染队列统计卡片
  const renderQueueStats = () => {
    if (!queueStats) return null;

    const totalTasks = queueStats.total_tasks;
    const completionRate = totalTasks > 0 ? (queueStats.completed_tasks / totalTasks) * 100 : 0;

    return (
      <Card
        title={
          <Space>
            <ClockCircleOutlined />
            队列状态
          </Space>
        }
        loading={loading}
      >
        <Row gutter={16}>
          <Col span={6}>
            <Statistic
              title="待处理"
              value={queueStats.pending_tasks}
              valueStyle={{ color: '#faad14' }}
              prefix={<ClockCircleOutlined />}
            />
          </Col>
          <Col span={6}>
            <Statistic
              title="处理中"
              value={queueStats.processing_tasks}
              valueStyle={{ color: '#1890ff' }}
              prefix={<ThunderboltOutlined />}
            />
          </Col>
          <Col span={6}>
            <Statistic
              title="已完成"
              value={queueStats.completed_tasks}
              valueStyle={{ color: '#52c41a' }}
              prefix={<CheckCircleOutlined />}
            />
          </Col>
          <Col span={6}>
            <Statistic
              title="失败"
              value={queueStats.failed_tasks}
              valueStyle={{ color: '#ff4d4f' }}
              prefix={<ExclamationCircleOutlined />}
            />
          </Col>
        </Row>

        <Divider />

        <Row gutter={16}>
          <Col span={12}>
            <div>
              <div style={{ marginBottom: 8 }}>完成率</div>
              <Progress
                percent={Math.round(completionRate)}
                status={completionRate === 100 ? 'success' : 'active'}
              />
            </div>
          </Col>
          <Col span={12}>
            <Descriptions size="small" column={1}>
              <Descriptions.Item label="工作线程">
                <Space>
                  <TeamOutlined />
                  {queueStats.worker_count}
                </Space>
              </Descriptions.Item>
              <Descriptions.Item label="平均等待时间">
                {ProcessingUtils.formatDuration(queueStats.average_wait_time)}
              </Descriptions.Item>
            </Descriptions>
          </Col>
        </Row>
      </Card>
    );
  };

  // 渲染处理统计卡片
  const renderProcessingStats = () => {
    if (!processingStats) return null;

    return (
      <Card
        title={
          <Space>
            <FileTextOutlined />
            处理统计
          </Space>
        }
        loading={loading}
      >
        <Row gutter={16}>
          <Col span={8}>
            <Statistic
              title="总文档数"
              value={processingStats.total_documents}
              prefix={<FileTextOutlined />}
            />
          </Col>
          <Col span={8}>
            <Statistic
              title="已处理"
              value={processingStats.processed_documents}
              valueStyle={{ color: '#52c41a' }}
            />
          </Col>
          <Col span={8}>
            <Statistic
              title="处理失败"
              value={processingStats.failed_documents}
              valueStyle={{ color: '#ff4d4f' }}
            />
          </Col>
        </Row>

        <Divider />

        <Row gutter={16}>
          <Col span={12}>
            <div>
              <div style={{ marginBottom: 8 }}>处理成功率</div>
              <Progress
                percent={Math.round(processingStats.processing_rate)}
                status={processingStats.processing_rate >= 90 ? 'success' : 'normal'}
              />
            </div>
          </Col>
          <Col span={12}>
            <Descriptions size="small" column={1}>
              <Descriptions.Item label="平均处理时间">
                {ProcessingUtils.formatDuration(processingStats.average_process_time)}
              </Descriptions.Item>
              <Descriptions.Item label="总处理大小">
                {ProcessingUtils.formatFileSize(processingStats.total_processed_size)}
              </Descriptions.Item>
              <Descriptions.Item label="处理吞吐量">
                {ProcessingUtils.formatFileSize(processingStats.processing_throughput)}/s
              </Descriptions.Item>
            </Descriptions>
          </Col>
        </Row>
      </Card>
    );
  };

  // 渲染支持格式卡片
  const renderSupportedFormats = () => {
    if (!supportedFormats) return null;

    return (
      <Card
        title="支持的文档格式"
        loading={loading}
      >
        <Space wrap>
          {supportedFormats.formats.map(format => (
            <Tag key={format} color="blue">
              .{format.toUpperCase()}
            </Tag>
          ))}
        </Space>
        <div style={{ marginTop: 16, color: '#666' }}>
          共支持 {supportedFormats.count} 种文档格式
        </div>
      </Card>
    );
  };

  return (
    <div>
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <h2>文档处理仪表板</h2>
        <Button
          icon={<ReloadOutlined />}
          onClick={loadStats}
          loading={loading}
        >
          刷新数据
        </Button>
      </div>

      {error && (
        <Alert
          message="加载统计数据失败"
          description={error}
          type="error"
          showIcon
          style={{ marginBottom: 16 }}
        />
      )}

      <Row gutter={[16, 16]}>
        <Col span={24}>
          {renderQueueStats()}
        </Col>
        
        <Col span={16}>
          {renderProcessingStats()}
        </Col>
        
        <Col span={8}>
          {renderSupportedFormats()}
        </Col>
      </Row>
    </div>
  );
};

export default ProcessingDashboard;