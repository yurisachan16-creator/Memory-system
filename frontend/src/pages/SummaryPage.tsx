import { App, Card, Col, Empty, Row, Space, Statistic, Tag, Typography } from "antd";
import { useEffect, useState } from "react";

import { getSummary } from "../api/memories";
import { MemorySnippetCard } from "../components/MemorySnippetCard";
import { SectionHeader } from "../components/SectionHeader";
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
          <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="No items yet" />
        )}
      </Space>
    </Card>
  );
}

export function SummaryPage() {
  const { message } = App.useApp();
  const { userId } = useUser();
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
          void message.error(error instanceof Error ? error.message : "Failed to load summary");
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
  }, [message, userId]);

  const { summary, cached } = data;

  return (
    <div className="page-stack">
      <Card className="hero-card" loading={loading}>
        <SectionHeader
          title="Memory Summary"
          subtitle="A dashboard view of the active user's strongest preferences, goals, background, and recent changes."
          extra={cached ? <Tag color="success">Cached summary</Tag> : <Tag color="default">Fresh summary</Tag>}
        />

        {userId ? (
          <div className="stats-grid">
            <Card className="surface-card stat-card">
              <Statistic title="Preferences" value={summary.preferences.length} />
            </Card>
            <Card className="surface-card stat-card">
              <Statistic title="Goals" value={summary.goals.length} />
            </Card>
            <Card className="surface-card stat-card">
              <Statistic title="Background" value={summary.background.length} />
            </Card>
            <Card className="surface-card stat-card">
              <Statistic title="Recent" value={summary.recent.length} />
            </Card>
          </div>
        ) : (
          <Empty description="Set a user_id in the top bar to load a summary dashboard." />
        )}
      </Card>

      {userId && (
        <Row gutter={[18, 18]}>
          <Col xs={24} xl={12}>
            <SummaryBucket
              title="Preferences"
              subtitle="Importance >= 3 memories that capture steady user tastes."
              items={summary.preferences}
            />
          </Col>
          <Col xs={24} xl={12}>
            <SummaryBucket title="Goals" subtitle="Latest goal-oriented memories." items={summary.goals} />
          </Col>
          <Col xs={24} xl={12}>
            <SummaryBucket
              title="Background"
              subtitle="Identity memories with the highest signal."
              items={summary.background}
            />
          </Col>
          <Col xs={24} xl={12}>
            <SummaryBucket title="Recent" subtitle="Memories created in the last 7 days." items={summary.recent} />
          </Col>
        </Row>
      )}
    </div>
  );
}
