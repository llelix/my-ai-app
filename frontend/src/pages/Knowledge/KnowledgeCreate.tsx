import React, { useEffect, useState } from 'react';
import {
  Card,
  Form,
  Input,
  Select,
  Switch,
  Button,
  message,
  Space,
  Typography,
  Divider,
} from 'antd';
import { useNavigate } from 'react-router-dom';
import { knowledgeService, categoryService, tagService } from '../../services';
import type { Category, Tag as TagType, CreateKnowledgeRequest } from '../../types';
import { DIFFICULTY_OPTIONS } from '../../types';

interface FormValues {
  title: string;
  category_id: number;
  tags?: string[];
  content: string;
  summary?: string;
  author?: string;
  source?: string;
  language?: string;
  difficulty?: 'easy' | 'medium' | 'hard';
  keywords?: string;
  is_published?: boolean;
}

const { Title } = Typography;
const { Option } = Select;
const { TextArea } = Input;

const KnowledgeCreate: React.FC = () => {
  const [form] = Form.useForm();
  const navigate = useNavigate();
  const [categories, setCategories] = useState<Category[]>([]);
  const [tags, setTags] = useState<TagType[]>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const [categoryRes, tagRes] = await Promise.all([
          categoryService.getCategories(),
          tagService.getTags(),
        ]);
        setCategories(categoryRes.data || []);
        setTags(tagRes.data || []);
      } catch (error) {
        message.error('加载分类和标签失败');
      }
    };
    fetchData();
  }, []);

  const onFinish = async (values: FormValues) => {
    setLoading(true);
    try {
      const createRequest: CreateKnowledgeRequest = {
        title: values.title,
        content: values.content,
        summary: values.summary,
        category_id: values.category_id,
        // Ensure tags are an array of strings (names)
        tags: values.tags || [], 
        metadata: {
          author: values.author,
          source: values.source,
          language: values.language,
          difficulty: values.difficulty,
          keywords: values.keywords,
        },
        is_published: values.is_published || false,
      };

      const response = await knowledgeService.createKnowledge(createRequest);
      message.success('知识创建成功');
      navigate(`/knowledge/${response.data?.id}`); // Navigate to detail page
    } catch (_error) {
      console.error('Failed to create knowledge:', _error);
      message.error('知识创建失败');
    } finally {
      setLoading(false);
    }
  };

  const handleCancel = () => {
    navigate('/knowledge');
  };

  return (
    <Card>
      <Title level={2}>创建知识</Title>
      <Form
        form={form}
        layout="vertical"
        onFinish={onFinish}
        initialValues={{ is_published: false }}
      >
        <Form.Item
          name="title"
          label="标题"
          rules={[{ required: true, message: '请输入知识标题' }]}
        >
          <Input placeholder="输入知识标题" />
        </Form.Item>

        <Form.Item
          name="category_id"
          label="分类"
          rules={[{ required: true, message: '请选择知识分类' }]}
        >
          <Select placeholder="选择分类">
            {categories.map((category) => (
              <Option key={category.id} value={category.id}>
                {category.name}
              </Option>
            ))}
          </Select>
        </Form.Item>

        <Form.Item name="tags" label="标签">
          <Select
            mode="multiple"
            placeholder="选择或输入标签"
            filterOption={(input, option) =>
              String(option?.children).toLowerCase().indexOf(input.toLowerCase()) >= 0
            }
          >
            {tags.map((tag) => (
              <Option key={tag.id} value={tag.name}>
                {tag.name}
              </Option>
            ))}
          </Select>
        </Form.Item>

        <Form.Item
          name="content"
          label="内容"
          rules={[{ required: true, message: '请输入知识内容' }]}
        >
          <TextArea rows={10} placeholder="输入知识的详细内容" />
        </Form.Item>

        <Form.Item name="summary" label="摘要">
          <TextArea rows={4} placeholder="输入知识摘要 (可选)" />
        </Form.Item>

        <Divider>元数据 (可选)</Divider>

        <Form.Item name="author" label="作者">
          <Input placeholder="作者名称" />
        </Form.Item>

        <Form.Item name="source" label="来源">
          <Input placeholder="知识来源" />
        </Form.Item>

        <Form.Item name="language" label="语言">
          <Input placeholder="知识语言，例如：中文，English" />
        </Form.Item>

        <Form.Item name="difficulty" label="难度">
          <Select placeholder="选择难度级别">
            {DIFFICULTY_OPTIONS.map((option) => (
              <Option key={option.value} value={option.value}>
                {option.label}
              </Option>
            ))}
          </Select>
        </Form.Item>

        <Form.Item name="keywords" label="关键词">
          <Input placeholder="用逗号分隔关键词" />
        </Form.Item>

        <Form.Item name="is_published" label="发布" valuePropName="checked">
          <Switch />
        </Form.Item>

        <Form.Item>
          <Space>
            <Button type="primary" htmlType="submit" loading={loading}>
              创建
            </Button>
            <Button onClick={handleCancel}>取消</Button>
          </Space>
        </Form.Item>
      </Form>
    </Card>
  );
};

export default KnowledgeCreate;