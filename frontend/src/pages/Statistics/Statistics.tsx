import React from 'react';
import { Card, Typography, Empty } from 'antd';

const { Title } = Typography;

const Statistics: React.FC = () => {
  return (
    <Card>
      <Title level={2}>统计分析</Title>
      <Empty description="功能开发中..." />
    </Card>
  );
};

export default Statistics;