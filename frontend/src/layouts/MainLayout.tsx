import React, { useState } from 'react';
import {
  Layout,
  Menu,
  Avatar,
  Dropdown,
  Button,
  Drawer,
  Space,
  Badge,
  Tooltip,
} from 'antd';
import {
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  BookOutlined,
  TagsOutlined,
  FolderOutlined,
  RobotOutlined,
  BarChartOutlined,
  SettingOutlined,
  UserOutlined,
  LogoutOutlined,
  BellOutlined,
  SunOutlined,
  MoonOutlined,
  FileTextOutlined,
} from '@ant-design/icons';
import { useNavigate, useLocation, Outlet } from 'react-router-dom';
import { useThemeStore } from '../store';
import { ROUTES } from '../utils/constants';
import { theme } from 'antd';

const { Header, Sider, Content } = Layout;

const menuItems = [
  {
    key: ROUTES.HOME,
    icon: <BookOutlined />,
    label: '知识库',
  },
  {
    key: ROUTES.CATEGORIES,
    icon: <FolderOutlined />,
    label: '分类管理',
  },
  {
    key: ROUTES.TAGS,
    icon: <TagsOutlined />,
    label: '标签管理',
  },
  {
    key: '/documents',
    icon: <FileTextOutlined />,
    label: '文档管理',
  },
  {
    key: ROUTES.AI_CHAT,
    icon: <RobotOutlined />,
    label: 'AI对话',
  },
  {
    key: ROUTES.STATISTICS,
    icon: <BarChartOutlined />,
    label: '统计分析',
  },
  {
    key: ROUTES.SETTINGS,
    icon: <SettingOutlined />,
    label: '系统设置',
  },
];

export const MainLayout: React.FC = () => {
  const [collapsed, setCollapsed] = useState(false);
  const [mobileDrawerVisible, setMobileDrawerVisible] = useState(false);
  const navigate = useNavigate();
  const location = useLocation();
  const { theme: themeMode, toggleTheme } = useThemeStore();
  const { token } = theme.useToken();

  const isMobile = window.innerWidth < 768;

  const handleMenuClick = ({ key }: { key: string }) => {
    navigate(key);
    if (isMobile) {
      setMobileDrawerVisible(false);
    }
  };

  const handleUserMenuClick = ({ key }: { key: string }) => {
    switch (key) {
      case 'profile':
        navigate('/profile');
        break;
      case 'settings':
        navigate(ROUTES.SETTINGS);
        break;
      case 'logout':
        // 这里实现登出逻辑
        console.log('登出');
        break;
    }
  };

  const userMenuItems = [
    {
      key: 'profile',
      icon: <UserOutlined />,
      label: '个人信息',
    },
    {
      key: 'settings',
      icon: <SettingOutlined />,
      label: '设置',
    },
    {
      key: 'divider-1',
      type: 'divider' as const,
    },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: '退出登录',
      danger: true,
    },
  ];

  const SideMenu = () => (
    <Menu
      theme={themeMode === 'dark' ? 'dark' : 'light'}
      mode="inline"
      selectedKeys={[location.pathname]}
      items={menuItems}
      onClick={handleMenuClick}
      style={{ border: 'none' }}
    />
  );

  const HeaderContent = () => (
    <div
      style={{
        padding: '0 16px',
        background: token.colorBgContainer,
        borderBottom: `1px solid ${token.colorBorderSecondary}`,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
      }}
    >
      <div style={{ display: 'flex', alignItems: 'center' }}>
        {isMobile ? (
          <Button
            type="text"
            icon={
              mobileDrawerVisible ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />
            }
            onClick={() => setMobileDrawerVisible(!mobileDrawerVisible)}
            style={{ marginRight: 16 }}
          />
        ) : (
          <Button
            type="text"
            icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
            onClick={() => setCollapsed(!collapsed)}
            style={{ marginRight: 16 }}
          />
        )}
        <div style={{ display: 'flex', alignItems: 'center' }}>
          <BookOutlined style={{ fontSize: 20, color: token.colorPrimary, marginRight: 8 }} />
          <h1 style={{ margin: 0, fontSize: 18, fontWeight: 600 }}>
            AI 知识库
          </h1>
        </div>
      </div>

      <Space size="middle">
        <Tooltip title={themeMode === 'light' ? '切换深色模式' : '切换浅色模式'}>
          <Button
            type="text"
            icon={themeMode === 'light' ? <MoonOutlined /> : <SunOutlined />}
            onClick={toggleTheme}
          />
        </Tooltip>

        <Badge count={0} showZero={false}>
          <Button type="text" icon={<BellOutlined />} />
        </Badge>

        <Dropdown
          menu={{
            items: userMenuItems,
            onClick: handleUserMenuClick,
          }}
          placement="bottomRight"
          arrow
        >
          <Space style={{ cursor: 'pointer' }}>
            <Avatar
              size="small"
              icon={<UserOutlined />}
              style={{ backgroundColor: token.colorPrimary }}
            />
            <span style={{ fontSize: 14 }}>浮浮酱</span>
          </Space>
        </Dropdown>
      </Space>
    </div>
  );

  return (
    <Layout style={{ minHeight: '100vh' }}>
      {!isMobile && (
        <Sider
          trigger={null}
          collapsible
          collapsed={collapsed}
          style={{
            background: token.colorBgContainer,
            borderRight: `1px solid ${token.colorBorderSecondary}`,
          }}
          width={240}
          collapsedWidth={80}
        >
          <div
            style={{
              height: 64,
              display: 'flex',
              alignItems: 'center',
              justifyContent: collapsed ? 'center' : 'flex-start',
              padding: collapsed ? 0 : '0 16px',
              borderBottom: `1px solid ${token.colorBorderSecondary}`,
            }}
          >
            {collapsed ? (
              <BookOutlined style={{ fontSize: 20, color: token.colorPrimary }} />
            ) : (
              <div style={{ display: 'flex', alignItems: 'center' }}>
                <BookOutlined
                  style={{ fontSize: 24, color: token.colorPrimary, marginRight: 8 }}
                />
                <span style={{ fontSize: 16, fontWeight: 600 }}>AI 知识库</span>
              </div>
            )}
          </div>
          <SideMenu />
        </Sider>
      )}

      {isMobile && (
        <Drawer
          title="AI 知识库"
          placement="left"
          onClose={() => setMobileDrawerVisible(false)}
          open={mobileDrawerVisible}
          bodyStyle={{ padding: 0 }}
          width={240}
        >
          <SideMenu />
        </Drawer>
      )}

      <Layout>
        <Header>
          <HeaderContent />
        </Header>
        <Content>
          <div
            style={{
              margin: '16px',
              padding: '24px',
              background: token.colorBgContainer,
              borderRadius: token.borderRadius,
              minHeight: 'calc(100vh - 120px)',
            }}
          >
            <Outlet />
          </div>
        </Content>
      </Layout>
    </Layout>
  );
};

export default MainLayout;