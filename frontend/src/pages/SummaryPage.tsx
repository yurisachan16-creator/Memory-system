import { App, Card, Col, Empty, Row, Space, Statistic, Tag, Typography } from "antd";
import { useEffect, useState } from "react";

import { getSummary } from "../api/memories";
import { MemorySnippetCard } from "../components/MemorySnippetCard";
import { SectionHeader } from "../components/SectionHeader";
import { useLanguage } from "../context/LanguageContext";
import { useUser } from "../context/UserContext";
import type { Memory, SummaryPayload } from "../types/memory";

const initialSummary: SummaryPayload = {
  summary: {
    preferences: [],
    goals: [],
    background: [],
    recent: []
  },
  cached: false
};

function SummaryBucket({ title, subtitle, items }: { title: string; subtitle: string; items: Memory[] }) {
  const { t } = useLanguage();

  return (
    <Card className="surface-card summary-panel">
      <Space direction="vertical" size={6} className="full-width">
        <Typography.Title level={4} className="summary-panel-title">
          {title}
        </Typography.Title>
        <Typography.Paragraph className="summary-panel-subtitle">{subtitle}</Typography.Paragraph>
        {items.length > 0 ? (
          <Space direction="vertical" size={12} className="full-width">
            {items.map((memory) => (
              <MemorySnippetCard key={memory.id} memory={memory} />
            ))}
          </Space>
        ) : (
          <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={t("summary.noItems")} />
        )}
      </Space>
    </Card>
  );
}

export function SummaryPage() {
  const { message } = App.useApp();
  const { userId } = useUser();
  const { t } = useLanguage();
  const [loading, setLoading] = useState(false);
  const [data, setData] = useState<SummaryPayload>(initialSummary);

  useEffect(() => {
    if (!userId) {
      setData(initialSummary);
      return;
    }

    let cancelled = false;

    async function load() {
      setLoading(true);
      try {
        const summary = await getSummary(userId);
        if (!cancelled) {
          setData(summary);
        }
      } catch (error) {
        if (!cancelled) {
          void message.error(error instanceof Error ? error.message : t("summary.loadError"));
        }
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    }

    void load();

    return () => {
      cancelled = true;
    };
  }, [message, t, userId]);

  const { summary, cached } = data;

  return (
    <div className="page-stack">
      <Card className="hero-card" loading={loading}>
        <SectionHeader
          title={t("summary.pageTitle")}
          subtitle={t("summary.pageSubtitle")}
          extra={
            cached
              ? <Tag color="success">{t("summary.cachedTag")}</Tag>
              : <Tag color="default">{t("summary.freshTag")}</Tag>
          }
        />

        {userId ? (
          <div className="stats-grid">
            <Card className="surface-card stat-card">
              <Statistic title={t("summary.preferences")} value={summary.preferences.length} />
            </Card>
            <Card className="surface-card stat-card">
              <Statistic title={t("summary.goals")} value={summary.goals.length} />
            </Card>
            <Card className="surface-card stat-card">
              <Statistic title={t("summary.background")} value={summary.background.length} />
            </Card>
            <Card className="surface-card stat-card">
              <Statistic title={t("summary.recent")} value={summary.recent.length} />
            </Card>
          </div>
        ) : (
          <Empty description={t("summary.emptyHint")} />
        )}
      </Card>

      {userId && (
        <Row gutter={[18, 18]}>
          <Col xs={24} xl={12}>
            <SummaryBucket
              title={t("summary.preferences")}
              subtitle={t("summary.prefSubtitle")}
              items={summary.preferences}
            />
          </Col>
          <Col xs={24} xl={12}>
            <SummaryBucket
              title={t("summary.goals")}
              subtitle={t("summary.goalsSubtitle")}
              items={summary.goals}
            />
          </Col>
          <Col xs={24} xl={12}>
            <SummaryBucket
              title={t("summary.background")}
              subtitle={t("summary.bgSubtitle")}
              items={summary.background}
            />
          </Col>
          <Col xs={24} xl={12}>
            <SummaryBucket
              title={t("summary.recent")}
              subtitle={t("summary.recentSubtitle")}
              items={summary.recent}
            />
          </Col>
        </Row>
      )}
    </div>
  );
}
