import React from 'react';
import { Card, Typography, Empty } from 'antd';

const { Title } = Typography;

const CategoryList: React.FC = () => {
  return (
    <Card>
      <Title level={2}>分类管理</Title>
      <Empty description="功能开发中..." />
    </Card>
  );
};

export default CategoryList;