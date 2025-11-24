import React from 'react';
import { Card, Typography, Empty } from 'antd';

const { Title } = Typography;

const KnowledgeEdit: React.FC = () => {
  return (
    <Card>
      <Title level={2}>编辑知识</Title>
      <Empty description="功能开发中..." />
    </Card>
  );
};

export default KnowledgeEdit;