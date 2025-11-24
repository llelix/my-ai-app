import React, { useState, useEffect, useRef } from 'react';
import {
  Card,
  Input,
  Button,
  Select,
  Space,
  message,
  Drawer,
  Slider,
  Divider,
  Tag,
  Empty,
  Spin,
  Tooltip,
  Avatar,
  Row,
  Col,
  Popconfirm,
} from 'antd';
import {
  SendOutlined,
  RobotOutlined,
  UserOutlined,
  SettingOutlined,
  HistoryOutlined,
  ClearOutlined,
  CopyOutlined,
  LikeOutlined,
  DislikeOutlined,
} from '@ant-design/icons';
import ReactMarkdown from 'react-markdown';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { oneLight } from 'react-syntax-highlighter/dist/esm/styles/prism';
import { useAIStore, useKnowledgeStore } from '../../store';
import { aiService } from '../../services';
import { copyToClipboard } from '../../utils/helpers';

const { TextArea } = Input;
const { Option } = Select;

interface Message {
  id: string;
  role: 'user' | 'assistant';
  content: string;
  timestamp: string;
  related_knowledges?: any[];
}

const AIChat: React.FC = () => {
  const [inputValue, setInputValue] = useState('');
  const [messages, setMessages] = useState<Message[]>([]);
  const [isQuerying, setIsQuerying] = useState(false);
  const [showSettings, setShowSettings] = useState(false);
  const [showHistory, setShowHistory] = useState(false);
  const [queryHistory, setQueryHistory] = useState<any[]>([]);
  const [feedbackStates, setFeedbackStates] = useState<Record<string, 'up' | 'down'>>({});

  const {
    selectedModel,
    temperature,
    maxTokens,
    setSelectedModel,
    setTemperature,
    setMaxTokens,
    addToHistory,
  } = useAIStore();

  const { knowledges } = useKnowledgeStore();
  const messagesEndRef = useRef<HTMLDivElement>(null);

  const models = [
    { label: 'GPT-3.5 Turbo', value: 'gpt-3.5-turbo' },
    { label: 'GPT-4', value: 'gpt-4' },
    { label: '通义千问 Plus', value: 'qwen-plus' },
    { label: '通义千问 Max', value: 'qwen-max' },
    { label: 'Moonshot-8K', value: 'moonshot-v1-8k' },
    { label: 'Moonshot-32K', value: 'moonshot-v1-32k' },
    { label: '智谱清言', value: 'glm-3-turbo' },
    { label: 'Claude-3-Sonnet', value: 'claude-3-sonnet-20240229' },
  ];

  // 加载查询历史
  const loadHistory = async () => {
    try {
      const res = await aiService.getQueryHistory({ page: 1, page_size: 50 });
      setQueryHistory(res.data?.items || []);
    } catch (error) {
      message.error('加载历史记录失败');
    }
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  // 发送查询
  const handleSend = async () => {
    if (!inputValue.trim()) {
      message.warning('请输入查询内容');
      return;
    }

    const userMessage: Message = {
      id: Date.now().toString(),
      role: 'user',
      content: inputValue.trim(),
      timestamp: new Date().toISOString(),
    };

    setMessages(prev => [...prev, userMessage]);
    setInputValue('');
    setIsQuerying(true);

    try {
      const res = await aiService.query({
        query: inputValue.trim(),
        model: selectedModel,
        temperature,
        max_tokens: maxTokens,
      });

      const assistantMessage: Message = {
        id: (Date.now() + 1).toString(),
        role: 'assistant',
        content: res.data?.response || '抱歉，没有得到有效回复',
        timestamp: new Date().toISOString(),
        related_knowledges: res.data?.related_knowledges || [],
      };

      setMessages(prev => [...prev, assistantMessage]);

      // 添加到历史记录
      addToHistory({
        query: userMessage.content,
        response: assistantMessage.content,
        model: res.data?.model || 'unknown',
        timestamp: assistantMessage.timestamp,
        related_knowledges: res.data?.related_knowledges || [],
      });

    } catch (error) {
      message.error('查询失败，请稍后重试');

      const errorMessage: Message = {
        id: (Date.now() + 1).toString(),
        role: 'assistant',
        content: '抱歉，查询时出现了错误。请检查网络连接或稍后重试。',
        timestamp: new Date().toISOString(),
      };

      setMessages(prev => [...prev, errorMessage]);
    } finally {
      setIsQuerying(false);
    }
  };

  // 复制消息
  const handleCopy = async (content: string) => {
    const success = await copyToClipboard(content);
    if (success) {
      message.success('复制成功');
    } else {
      message.error('复制失败');
    }
  };

  // 提交反馈
  const handleFeedback = (messageId: string, feedback: 'up' | 'down') => {
    setFeedbackStates(prev => ({ ...prev, [messageId]: feedback }));
    message.success(feedback === 'up' ? '感谢您的反馈！' : '我们会改进回答质量');
  };

  // 清空对话
  const handleClear = () => {
    setMessages([]);
    setFeedbackStates({});
  };

  // 历史记录项点击
  const handleHistoryItemClick = (item: any) => {
    setInputValue(item.query);
    setShowHistory(false);
  };

  // Markdown渲染器配置
  const markdownComponents = {
    code({ node, inline, className, children, ...props }: any) {
      const match = /language-(\w+)/.exec(className || '');
      return !inline && match ? (
        <SyntaxHighlighter
          style={oneLight}
          language={match[1]}
          PreTag="div"
          {...props}
        >
          {String(children).replace(/\n$/, '')}
        </SyntaxHighlighter>
      ) : (
        <code className={className} {...props}>
          {children}
        </code>
      );
    },
  };

  return (
    <div style={{ height: 'calc(100vh - 180px)' }}>
      <Row gutter={16} style={{ height: '100%' }}>
        {/* 左侧聊天区域 */}
        <Col xs={24} lg={16}>
          <Card
            title={
              <Space>
                <RobotOutlined />
                <span>AI 对话</span>
                <Tag color="blue">{selectedModel}</Tag>
              </Space>
            }
            extra={
              <Space>
                <Tooltip title="对话设置">
                  <Button
                    type="text"
                    icon={<SettingOutlined />}
                    onClick={() => setShowSettings(true)}
                  />
                </Tooltip>
                <Tooltip title="历史记录">
                  <Button
                    type="text"
                    icon={<HistoryOutlined />}
                    onClick={() => {
                      setShowHistory(true);
                      loadHistory();
                    }}
                  />
                </Tooltip>
                <Tooltip title="清空对话">
                  <Popconfirm
                    title="确定清空所有对话？"
                    onConfirm={handleClear}
                  >
                    <Button
                      type="text"
                      icon={<ClearOutlined />}
                    />
                  </Popconfirm>
                </Tooltip>
              </Space>
            }
            bodyStyle={{ padding: 0, height: 'calc(100% - 60px)' }}
          >
            {/* 消息列表 */}
            <div
              style={{
                height: 'calc(100% - 120px)',
                overflowY: 'auto',
                padding: '16px',
                backgroundColor: '#fafafa',
              }}
            >
              {messages.length === 0 ? (
                <Empty
                  description="开始与AI对话吧"
                  image={Empty.PRESENTED_IMAGE_SIMPLE}
                />
              ) : (
                messages.map((message) => (
                  <div
                    key={message.id}
                    style={{
                      marginBottom: 16,
                      display: 'flex',
                      justifyContent: message.role === 'user' ? 'flex-end' : 'flex-start',
                    }}
                  >
                    <div
                      style={{
                        maxWidth: '70%',
                        display: 'flex',
                        alignItems: 'flex-start',
                        gap: 8,
                        flexDirection: message.role === 'user' ? 'row-reverse' : 'row',
                      }}
                    >
                      <Avatar
                        icon={
                          message.role === 'user' ? <UserOutlined /> : <RobotOutlined />
                        }
                        style={{
                          backgroundColor: message.role === 'user' ? '#1890ff' : '#52c41a',
                        }}
                      />
                      <div>
                        <div
                          style={{
                            padding: '12px 16px',
                            borderRadius: 12,
                            backgroundColor: message.role === 'user' ? '#1890ff' : '#fff',
                            color: message.role === 'user' ? '#fff' : '#333',
                            boxShadow: '0 2px 8px rgba(0,0,0,0.1)',
                            marginBottom: 4,
                          }}
                        >
                          {message.role === 'user' ? (
                            <div>{message.content}</div>
                          ) : (
                            <div>
                              <ReactMarkdown
                                components={markdownComponents}
                              >
                                {message.content}
                              </ReactMarkdown>
                            </div>
                          )}
                        </div>

                        {/* 相关知识 */}
                        {message.related_knowledges && message.related_knowledges.length > 0 && (
                          <div style={{ margin: '8px 0' }}>
                            <div style={{ fontSize: 12, color: '#666', marginBottom: 4 }}>
                              相关知识：
                            </div>
                            <Space wrap>
                              {message.related_knowledges.map((knowledge) => (
                                <Tag
                                  key={knowledge.id}
                                  color="blue"
                                  style={{ cursor: 'pointer' }}
                                  onClick={() => window.open(`/knowledge/${knowledge.id}`)}
                                >
                                  {knowledge.title}
                                </Tag>
                              ))}
                            </Space>
                          </div>
                        )}

                        {/* 操作按钮 */}
                        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                          <span style={{ fontSize: 12, color: '#999' }}>
                            {new Date(message.timestamp).toLocaleTimeString()}
                          </span>
                          {message.role === 'assistant' && (
                            <Space size={4}>
                              <Tooltip title="复制">
                                <Button
                                  type="text"
                                  size="small"
                                  icon={<CopyOutlined />}
                                  onClick={() => handleCopy(message.content)}
                                />
                              </Tooltip>
                              <Tooltip title="有帮助">
                                <Button
                                  type="text"
                                  size="small"
                                  icon={<LikeOutlined/>}
                                  style={{
                                    color: feedbackStates[message.id] === 'up' ? '#1890ff' : '#666',
                                  }}
                                  onClick={() => handleFeedback(message.id, 'up')}
                                />
                              </Tooltip>
                              <Tooltip title="没帮助">
                                <Button
                                  type="text"
                                  size="small"
                                  icon={<DislikeOutlined/>}
                                  style={{
                                    color: feedbackStates[message.id] === 'down' ? '#ff4d4f' : '#666',
                                  }}
                                  onClick={() => handleFeedback(message.id, 'down')}
                                />
                              </Tooltip>
                            </Space>
                          )}
                        </div>
                      </div>
                    </div>
                  </div>
                ))
              )}
              {isQuerying && (
                <div style={{ display: 'flex', alignItems: 'flex-start', gap: 8 }}>
                  <Avatar icon={<RobotOutlined />} style={{ backgroundColor: '#52c41a' }} />
                  <div
                    style={{
                      padding: '12px 16px',
                      borderRadius: 12,
                      backgroundColor: '#fff',
                      boxShadow: '0 2px 8px rgba(0,0,0,0.1)',
                    }}
                  >
                    <Spin size="small" /> 思考中...
                  </div>
                </div>
              )}
              <div ref={messagesEndRef} />
            </div>

            {/* 输入区域 */}
            <div style={{ padding: '16px', borderTop: '1px solid #f0f0f0' }}>
              <Space.Compact style={{ width: '100%' }}>
                <TextArea
                  value={inputValue}
                  onChange={(e) => setInputValue(e.target.value)}
                  placeholder="请输入您的问题..."
                  autoSize={{ minRows: 2, maxRows: 6 }}
                  onPressEnter={(e) => {
                    if (!e.shiftKey) {
                      e.preventDefault();
                      handleSend();
                    }
                  }}
                />
                <Button
                  type="primary"
                  icon={<SendOutlined />}
                  onClick={handleSend}
                  loading={isQuerying}
                  disabled={!inputValue.trim()}
                >
                  发送
                </Button>
              </Space.Compact>
            </div>
          </Card>
        </Col>

        {/* 右侧知识推荐区域 */}
        <Col xs={0} lg={8}>
          <Card
            title="知识库推荐"
            size="small"
            bodyStyle={{ padding: '12px', height: 'calc(100vh - 220px)', overflow: 'hidden' }}
          >
            <div style={{
              height: '100%',
              overflowY: 'auto',
              paddingRight: '4px'
            }}>
              {knowledges.length === 0 ? (
                <Empty
                  description="暂无知识库内容"
                  image={Empty.PRESENTED_IMAGE_SIMPLE}
                  style={{ marginTop: '20px' }}
                />
              ) : (
                knowledges.slice(0, 10).map((knowledge) => (
                  <div
                    key={knowledge.id}
                    style={{
                      padding: '8px',
                      marginBottom: 8,
                      border: '1px solid #f0f0f0',
                      borderRadius: 6,
                      cursor: 'pointer',
                      transition: 'all 0.2s ease',
                      backgroundColor: '#fff',
                    }}
                    onClick={() => window.open(`/knowledge/${knowledge.id}`)}
                    onMouseEnter={(e) => {
                      e.currentTarget.style.backgroundColor = '#f6f8fa';
                      e.currentTarget.style.borderColor = '#d9d9d9';
                    }}
                    onMouseLeave={(e) => {
                      e.currentTarget.style.backgroundColor = '#fff';
                      e.currentTarget.style.borderColor = '#f0f0f0';
                    }}
                  >
                    <div style={{
                      fontWeight: 500,
                      marginBottom: 4,
                      fontSize: '14px',
                      lineHeight: '1.4'
                    }}>
                      {knowledge.title}
                    </div>
                    <div style={{
                      fontSize: 12,
                      color: '#666',
                      lineHeight: '1.4',
                      display: '-webkit-box',
                      WebkitLineClamp: 2,
                      WebkitBoxOrient: 'vertical',
                      overflow: 'hidden'
                    }}>
                      {knowledge.summary || '暂无摘要'}
                    </div>
                  </div>
                ))
              )}
            </div>
          </Card>
        </Col>
      </Row>

      {/* 设置抽屉 */}
      <Drawer
        title="对话设置"
        placement="right"
        onClose={() => setShowSettings(false)}
        open={showSettings}
      >
        <Space direction="vertical" style={{ width: '100%' }}>
          <div>
            <div style={{ marginBottom: 8 }}>模型选择</div>
            <Select
              value={selectedModel}
              onChange={setSelectedModel}
              style={{ width: '100%' }}
            >
              {models.map((model) => (
                <Option key={model.value} value={model.value}>
                  {model.label}
                </Option>
              ))}
            </Select>
          </div>

          <Divider />

          <div>
            <div style={{ marginBottom: 8 }}>
              创造性: {temperature}
            </div>
            <Slider
              min={0}
              max={1}
              step={0.1}
              value={temperature}
              onChange={setTemperature}
            />
          </div>

          <div>
            <div style={{ marginBottom: 8 }}>
              最大字数: {maxTokens}
            </div>
            <Slider
              min={100}
              max={4000}
              step={100}
              value={maxTokens}
              onChange={setMaxTokens}
            />
          </div>
        </Space>
      </Drawer>

      {/* 历史记录抽屉 */}
      <Drawer
        title="历史记录"
        placement="right"
        onClose={() => setShowHistory(false)}
        open={showHistory}
      >
        <div>
          {queryHistory.map((item, index) => (
            <div
              key={index}
              style={{
                padding: '12px',
                marginBottom: 8,
                border: '1px solid #f0f0f0',
                borderRadius: 6,
                cursor: 'pointer',
              }}
              onClick={() => handleHistoryItemClick(item)}
            >
              <div style={{ fontWeight: 500, marginBottom: 4 }}>
                {item.query}
              </div>
              <div style={{ fontSize: 12, color: '#666' }}>
                {item.model} · {new Date(item.created_at).toLocaleString()}
              </div>
            </div>
          ))}
        </div>
      </Drawer>
    </div>
  );
};

export default AIChat;