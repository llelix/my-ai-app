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
} from 'antd';
import { useNavigate } from 'react-router-dom';
import { knowledgeService, tagService } from '../../services';
import type { Tag as TagType, CreateKnowledgeRequest } from '../../types';

interface FormValues {
  title: string;
  tags?: string[];
  content: string;
  summary?: string;
  is_published?: boolean;
}

const { Title } = Typography;
const { Option } = Select;
const { TextArea } = Input;

const KnowledgeCreate: React.FC = () => {
  const [form] = Form.useForm();
  const navigate = useNavigate();
  const [tags, setTags] = useState<TagType[]>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const tagRes = await tagService.getTags();
        setTags(tagRes.data || []);
      } catch (error) {
        message.error('加载标签失败');
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
        tags: values.tags || [],
        is_published: values.is_published || false,
      };

      const response = await knowledgeService.createKnowledge(createRequest);
      message.success('知识创建成功');
      navigate(`/knowledge/${response.data?.id}`);
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