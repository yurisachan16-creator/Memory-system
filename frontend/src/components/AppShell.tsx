import { DashboardOutlined, SearchOutlined, TranslationOutlined, UnorderedListOutlined } from "@ant-design/icons";
import { Alert, Button, Input, Layout, Menu, Space, Tag, Typography } from "antd";
import type { MenuProps } from "antd";
import { Outlet, useLocation, useNavigate } from "react-router-dom";

import { useLanguage } from "../context/LanguageContext";
import { useUser } from "../context/UserContext";

const { Header, Content } = Layout;

export function AppShell() {
  const location = useLocation();
  const navigate = useNavigate();
  const { userId, draftUserId, setDraftUserId, confirmUserId } = useUser();
  const { t, toggleLang } = useLanguage();

  const navItems: MenuProps["items"] = [
    { key: "/memories", icon: <UnorderedListOutlined />, label: t("nav.memories") },
    { key: "/search", icon: <SearchOutlined />, label: t("nav.search") },
    { key: "/summary", icon: <DashboardOutlined />, label: t("nav.summary") }
  ];

  return (
    <Layout className="app-layout">
      <div className="ambient-orb ambient-orb-left" />
      <div className="ambient-orb ambient-orb-right" />
      <Header className="app-header">
        <div className="brand-block">
          <Typography.Title level={2} className="brand-title">
            {t("app.title")}
          </Typography.Title>
          <Typography.Paragraph className="brand-subtitle">
            {t("app.subtitle")}
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
              placeholder={t("header.userIdPlaceholder")}
              value={draftUserId}
              onChange={(event) => setDraftUserId(event.target.value)}
              onPressEnter={confirmUserId}
            />
            <Button type="primary" onClick={confirmUserId}>
              {t("header.apply")}
            </Button>
          </Space.Compact>

          {userId
            ? <Tag color="processing">{t("header.activeUser", { id: userId })}</Tag>
            : <Tag color="warning">{t("header.userNotSet")}</Tag>
          }

          <Button
            icon={<TranslationOutlined />}
            onClick={toggleLang}
            title="Switch language / 切换语言"
          >
            {t("header.langToggle")}
          </Button>
        </div>
      </Header>

      <Content className="app-content">
        {!userId && (
          <Alert
            className="global-alert"
            type="warning"
            showIcon
            message={t("header.noUser")}
          />
        )}
        <Outlet />
      </Content>
    </Layout>
  );
}
