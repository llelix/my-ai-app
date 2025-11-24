import React, { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
  Card,
  Button,
  Tag,
  Space,
  Divider,
  Typography,
  Spin,
  message,
  Breadcrumb,
} from 'antd';
import {
  EditOutlined,
  EyeOutlined,
  ShareAltOutlined,
  StarOutlined,
  StarFilled,
} from '@ant-design/icons';
import { knowledgeService } from '../../services';
import type { Knowledge } from '../../types';
import { formatDate } from '../../utils/helpers';

const { Title, Paragraph } = Typography;

const KnowledgeDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [knowledge, setKnowledge] = useState<Knowledge | null>(null);
  const [loading, setLoading] = useState(true);
  const [isFavorite, setIsFavorite] = useState(false);

  useEffect(() => {
    if (id) {
      loadKnowledge(parseInt(id));
    }
  }, [id]);

  const loadKnowledge = async (id: number) => {
    setLoading(true);
    try {
      const res = await knowledgeService.getKnowledge(id);
      setKnowledge(res.data || null);
      // 增加查看次数
      await knowledgeService.incrementViewCount(id);
    } catch (error) {
      message.error('加载知识失败');
    } finally {
      setLoading(false);
    }
  };

  const handleShare = () => {
    const url = window.location.href;
    navigator.clipboard.writeText(url);
    message.success('链接已复制到剪贴板');
  };

  const toggleFavorite = () => {
    setIsFavorite(!isFavorite);
    message.success(isFavorite ? '已取消收藏' : '已添加收藏');
  };

  if (loading) {
    return (
      <div style={{ textAlign: 'center', padding: '50px' }}>
        <Spin size="large" />
      </div>
    );
  }

  if (!knowledge) {
    return (
      <div>
        <p>知识不存在</p>
      </div>
    );
  }

  return (
    <div>
      {/* 面包屑导航 */}
      <Breadcrumb style={{ marginBottom: 16 }}>
        <Breadcrumb.Item>知识库</Breadcrumb.Item>
        <Breadcrumb.Item>知识详情</Breadcrumb.Item>
        <Breadcrumb.Item>{knowledge.title}</Breadcrumb.Item>
      </Breadcrumb>

      <Card>
        {/* 标题和操作按钮 */}
        <div style={{ marginBottom: 24 }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
            <div style={{ flex: 1 }}>
              <Title level={2} style={{ margin: 0, marginBottom: 8 }}>
                {knowledge.title}
              </Title>
              <Space>
                <Tag color={knowledge.category?.color || 'default'}>
                  {knowledge.category?.name || '未分类'}
                </Tag>
                <Space size={4}>
                  <EyeOutlined />
                  <span>{knowledge.view_count}</span>
                </Space>
                <span style={{ color: '#666' }}>
                  {formatDate(knowledge.created_at)}
                </span>
              </Space>
            </div>
            <Space>
              <Button
                type="text"
                icon={isFavorite ? <StarFilled /> : <StarOutlined />}
                onClick={toggleFavorite}
                style={{ color: isFavorite ? '#faad14' : '#666' }}
              />
              <Button icon={<ShareAltOutlined />} onClick={handleShare}>
                分享
              </Button>
              <Button icon={<EditOutlined />} onClick={() => navigate(`/knowledge/${knowledge.id}/edit`)}>
                编辑
              </Button>
            </Space>
          </div>
        </div>

        <Divider />

        {/* 元信息 */}
        <div style={{ marginBottom: 24 }}>
          <Space wrap>
            {knowledge.tags.map((tag) => (
              <Tag key={tag.id} color={tag.color}>
                {tag.name}
              </Tag>
            ))}
          </Space>
          {knowledge.metadata.author && (
            <div style={{ marginTop: 8, color: '#666' }}>
              作者: {knowledge.metadata.author}
            </div>
          )}
        </div>

        {/* 摘要 */}
        {knowledge.summary && (
          <div style={{ marginBottom: 24 }}>
            <Title level={4}>摘要</Title>
            <Paragraph style={{ backgroundColor: '#f6f8fa', padding: '12px', borderRadius: 6 }}>
              {knowledge.summary}
            </Paragraph>
          </div>
        )}

        {/* 内容 */}
        <div>
          <Title level={4}>内容</Title>
          <div>
            {knowledge.content.split('\n').map((paragraph, index) => (
              <Paragraph key={index}>
                {paragraph}
              </Paragraph>
            ))}
          </div>
        </div>

        {/* 额外信息 */}
        {(knowledge.metadata.source || knowledge.metadata.keywords) && (
          <>
            <Divider />
            <div>
              <Title level={5}>元数据</Title>
              {knowledge.metadata.source && (
                <div style={{ marginBottom: 8 }}>
                  <strong>来源:</strong> {knowledge.metadata.source}
                </div>
              )}
              {knowledge.metadata.keywords && (
                <div style={{ marginBottom: 8 }}>
                  <strong>关键词:</strong> {knowledge.metadata.keywords}
                </div>
              )}
              {knowledge.metadata.difficulty && (
                <div style={{ marginBottom: 8 }}>
                  <strong>难度:</strong> {knowledge.metadata.difficulty}
                </div>
              )}
            </div>
          </>
        )}
      </Card>
    </div>
  );
};

export default KnowledgeDetail;