import React, { useEffect, useState } from 'react';
import {
  Card,
  Table,
  Button,
  Input,
  Select,
  Space,
  Tag,
  Tooltip,
  Modal,
  message,
  Popconfirm,
  Badge,
  Avatar,
} from 'antd';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  EyeOutlined,
  ExportOutlined,
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import { useNavigate } from 'react-router-dom';
import { useKnowledgeStore } from '../../store';
import { knowledgeService, tagService } from '../../services';
import type { Knowledge, Tag as TagType } from '../../types';
import { formatDate, truncateText } from '../../utils/helpers';

const { Search } = Input;
const { Option } = Select;

const KnowledgeList: React.FC = () => {
  const navigate = useNavigate();
  const {
    knowledges,
    tags,
    pagination,
    filters,
    loading,
    setKnowledges,
    setTags,
    setPagination,
    setFilters,
    setLoading,
  } = useKnowledgeStore();

  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([]);

  // 加载数据
  const loadData = async (params?: any) => {
    setLoading(true);
    try {
      const [knowledgeRes, tagRes] = await Promise.all([
        knowledgeService.getKnowledges({
          page: pagination.page,
          page_size: pagination.page_size,
          search: filters.search,
          tag_id: filters.tag_id || undefined,
          ...params,
        }),
        tagService.getTags(),
      ]);

      setKnowledges(knowledgeRes.data?.items || []);
      setPagination({
        page: knowledgeRes.data?.page || 1,
        page_size: knowledgeRes.data?.page_size || 10,
        total: knowledgeRes.data?.total || 0,
        total_pages: knowledgeRes.data?.total_pages || 0,
      });
      setTags(tagRes.data || []);
    } catch (error) {
      message.error('加载数据失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, [pagination.page, pagination.page_size]);

  // 搜索处理
  const handleSearch = (value: string) => {
    setFilters({ search: value });
    setPagination({ page: 1 });
    loadData();
  };

  // 筛选处理
  const handleFilter = (key: string, value: any) => {
    setFilters({ [key]: value });
    setPagination({ page: 1 });
    loadData();
  };

  // 删除知识
  const handleDelete = async (id: number) => {
    try {
      await knowledgeService.deleteKnowledge(id);
      message.success('删除成功');
      loadData();
    } catch (error) {
      message.error('删除失败');
    }
  };

  // 批量删除
  const handleBatchDelete = async () => {
    if (selectedRowKeys.length === 0) {
      message.warning('请选择要删除的项');
      return;
    }

    Modal.confirm({
      title: '批量删除',
      content: `确定要删除选中的 ${selectedRowKeys.length} 项吗？`,
      onOk: async () => {
        try {
          await knowledgeService.batchDelete(selectedRowKeys as number[]);
          message.success('批量删除成功');
          setSelectedRowKeys([]);
          loadData();
        } catch (error) {
          message.error('批量删除失败');
        }
      },
    });
  };

  // 导出功能
  const handleExport = () => {
    message.info('导出功能开发中...');
  };

  // 表格列定义
  const columns: ColumnsType<Knowledge> = [
    {
      title: '标题',
      dataIndex: 'title',
      key: 'title',
      width: 300,
      render: (text, record) => (
        <div>
          <a
            onClick={() => navigate(`/knowledge/${record.id}`)}
            style={{ fontWeight: 500, color: '#1890ff' }}
          >
            {text}
          </a>
          <div style={{ fontSize: 12, color: '#666', marginTop: 4 }}>
            {truncateText(record.content, 50)}
          </div>
        </div>
      ),
    },
    {
      title: '标签',
      dataIndex: 'tags',
      key: 'tags',
      width: 200,
      render: (tags: TagType[]) => (
        <div style={{ display: 'flex', flexWrap: 'wrap', gap: 4 }}>
          {tags.slice(0, 3).map((tag) => (
            <Tag key={tag.id} color={tag.color} style={{ fontSize: '12px' }}>
              {tag.name}
            </Tag>
          ))}
          {tags.length > 3 && (
            <Tag style={{ fontSize: '12px' }}>+{tags.length - 3}</Tag>
          )}
        </div>
      ),
    },
    {
      title: '作者',
      dataIndex: ['metadata', 'author'],
      key: 'author',
      width: 100,
      render: (author) => (
        <div style={{ display: 'flex', alignItems: 'center' }}>
          <Avatar size="small" style={{ marginRight: 8 }}>
            {author?.charAt(0) || '未'}
          </Avatar>
          {author || '未知'}
        </div>
      ),
    },
    {
      title: '查看次数',
      dataIndex: 'view_count',
      key: 'view_count',
      width: 100,
      sorter: true,
      render: (count) => <Badge count={count} showZero style={{ backgroundColor: '#52c41a' }} />,
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 150,
      sorter: true,
      render: (date) => formatDate(date),
    },
    {
      title: '状态',
      dataIndex: 'is_published',
      key: 'is_published',
      width: 80,
      render: (published) => (
        <Badge
          status={published ? 'success' : 'warning'}
          text={published ? '已发布' : '草稿'}
        />
      ),
    },
    {
      title: '操作',
      key: 'actions',
      width: 120,
      fixed: 'right',
      render: (_, record) => (
        <Space>
          <Tooltip title="查看">
            <Button
              type="text"
              icon={<EyeOutlined />}
              onClick={() => navigate(`/knowledge/${record.id}`)}
            />
          </Tooltip>
          <Tooltip title="编辑">
            <Button
              type="text"
              icon={<EditOutlined />}
              onClick={() => navigate(`/knowledge/${record.id}/edit`)}
            />
          </Tooltip>
          <Popconfirm
            title="确定删除吗？"
            onConfirm={() => handleDelete(record.id)}
            okText="确定"
            cancelText="取消"
          >
            <Tooltip title="删除">
              <Button type="text" danger icon={<DeleteOutlined />} />
            </Tooltip>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  // 表格选择配置
  const rowSelection = {
    selectedRowKeys,
    onChange: (newSelectedRowKeys: React.Key[]) => {
      setSelectedRowKeys(newSelectedRowKeys);
    },
  };

  return (
    <div>
      <Card>
        {/* 页面标题和操作按钮 */}
        <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <div>
            <h2 style={{ margin: 0 }}>知识库管理</h2>
            <p style={{ margin: '4px 0 0 0', color: '#666' }}>
              共 {pagination.total} 条知识
            </p>
          </div>
          <Space>
            {selectedRowKeys.length > 0 && (
              <>
                <span>已选择 {selectedRowKeys.length} 项</span>
                <Button onClick={handleBatchDelete} danger>
                  批量删除
                </Button>
              </>
            )}
            <Button icon={<ExportOutlined />} onClick={handleExport}>
              导出
            </Button>
            <Button type="primary" icon={<PlusOutlined />} onClick={() => navigate('/knowledge/create')}>
              新建知识
            </Button>
          </Space>
        </div>

        {/* 搜索和筛选 */}
        <div style={{ marginBottom: 16, display: 'flex', gap: 12, flexWrap: 'wrap' }}>
          <Search
            placeholder="搜索知识标题、内容..."
            style={{ width: 300 }}
            onSearch={handleSearch}
            enterButton
            allowClear
          />
          <Select
            placeholder="选择标签"
            style={{ width: 150 }}
            value={filters.tag_id}
            onChange={(value) => handleFilter('tag_id', value)}
            allowClear
          >
            {tags.map((tag: TagType) => (
              <Option key={tag.id} value={tag.id}>
                {tag.name}
              </Option>
            ))}
          </Select>
        </div>

        {/* 表格 */}
        <Table
          columns={columns}
          dataSource={knowledges}
          rowKey="id"
          loading={loading}
          pagination={{
            current: pagination.page,
            pageSize: pagination.page_size,
            total: pagination.total,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total, range) =>
              `第 ${range[0]}-${range[1]} 条，共 ${total} 条`,
            pageSizeOptions: ['10', '20', '50', '100'],
            onChange: (page, pageSize) => {
              setPagination({ page, page_size: pageSize || 10 });
              loadData();
            },
          }}
          rowSelection={rowSelection}
          scroll={{ x: 1200 }}
        />
      </Card>
    </div>
  );
};

export default KnowledgeList;