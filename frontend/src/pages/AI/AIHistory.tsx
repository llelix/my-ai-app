import React from 'react';
import { Card, Typography, Empty } from 'antd';

const { Title } = Typography;

const AIHistory: React.FC = () => {
  return (
    <Card>
      <Title level={2}>查询历史</Title>
      <Empty description="功能开发中..." />
    </Card>
  );
};

export default AIHistory;