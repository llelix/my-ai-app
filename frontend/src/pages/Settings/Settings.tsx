import React from 'react';
import { Card, Typography, Empty } from 'antd';

const { Title } = Typography;

const Settings: React.FC = () => {
  return (
    <Card>
      <Title level={2}>系统设置</Title>
      <Empty description="功能开发中..." />
    </Card>
  );
};

export default Settings;