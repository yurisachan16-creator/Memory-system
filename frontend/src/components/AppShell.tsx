import { DashboardOutlined, SearchOutlined, UnorderedListOutlined } from "@ant-design/icons";
import { Alert, Button, Input, Layout, Menu, Space, Tag, Typography } from "antd";
import type { MenuProps } from "antd";
import { Outlet, useLocation, useNavigate } from "react-router-dom";

import { useUser } from "../context/UserContext";

const { Header, Content } = Layout;

const navItems: MenuProps["items"] = [
  { key: "/memories", icon: <UnorderedListOutlined />, label: "Memories" },
  { key: "/search", icon: <SearchOutlined />, label: "Search" },
  { key: "/summary", icon: <DashboardOutlined />, label: "Summary" }
];

export function AppShell() {
  const location = useLocation();
  const navigate = useNavigate();
  const { userId, draftUserId, setDraftUserId, confirmUserId } = useUser();

  return (
    <Layout className="app-layout">
      <div className="ambient-orb ambient-orb-left" />
      <div className="ambient-orb ambient-orb-right" />
      <Header className="app-header">
        <div className="brand-block">
          <Typography.Title level={2} className="brand-title">
            Memory System Console
          </Typography.Title>
          <Typography.Paragraph className="brand-subtitle">
            A focused workspace for curating, searching, and summarizing user memory.
          </Typography.Paragraph>
        </div>

        <div className="header-controls">
          <Menu
            mode="horizontal"
            selectedKeys={[location.pathname]}
            items={navItems}
            onClick={({ key }) => navigate(key)}
            className="nav-menu"
          />

          <Space.Compact className="user-switcher">
            <Input
              aria-label="Current user id"
              placeholder="Enter active user_id"
              value={draftUserId}
              onChange={(event) => setDraftUserId(event.target.value)}
              onPressEnter={confirmUserId}
            />
            <Button type="primary" onClick={confirmUserId}>
              Apply
            </Button>
          </Space.Compact>

          {userId ? <Tag color="processing">Active user: {userId}</Tag> : <Tag color="warning">User not set</Tag>}
        </div>
      </Header>

      <Content className="app-content">
        {!userId && (
          <Alert
            className="global-alert"
            type="warning"
            showIcon
            message="Set a user_id from the top bar before interacting with memories."
          />
        )}
        <Outlet />
      </Content>
    </Layout>
  );
}
