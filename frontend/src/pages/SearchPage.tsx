import { SearchOutlined } from "@ant-design/icons";
import { App, Button, Card, Empty, Input, List, Space, Statistic, Tag, Typography } from "antd";
import { useEffect, useState } from "react";

import { searchMemories } from "../api/memories";
import { HighlightText } from "../components/HighlightText";
import { SectionHeader } from "../components/SectionHeader";
import { useLanguage } from "../context/LanguageContext";
import { useUser } from "../context/UserContext";
import { categoryKeys } from "../i18n";
import { formatDateTime, type SearchItem } from "../types/memory";

export function SearchPage() {
  const { message } = App.useApp();
  const { userId } = useUser();
  const { t } = useLanguage();
  const [query, setQuery] = useState("");
  const [activeQuery, setActiveQuery] = useState("");
  const [loading, setLoading] = useState(false);
  const [items, setItems] = useState<SearchItem[]>([]);
  const [cached, setCached] = useState(false);

  useEffect(() => {
    if (!userId || !activeQuery) {
      setItems([]);
      setCached(false);
      return;
    }

    let cancelled = false;

    async function load() {
      setLoading(true);
      try {
        const data = await searchMemories({
          user_id: userId,
          query: activeQuery,
          limit: 5
        });
        if (!cancelled) {
          setItems(data.items);
          setCached(data.cached);
        }
      } catch (error) {
        if (!cancelled) {
          void message.error(error instanceof Error ? error.message : t("search.searchError"));
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
  }, [activeQuery, message, t, userId]);

  const runSearch = () => {
    const nextQuery = query.trim();
    if (!userId) {
      void message.warning(t("search.noUserWarning"));
      return;
    }
    if (!nextQuery) {
      void message.warning(t("search.emptyQuery"));
      return;
    }
    setActiveQuery(nextQuery);
  };

  return (
    <div className="page-stack">
      <Card className="hero-card">
        <SectionHeader
          title={t("search.pageTitle")}
          subtitle={t("search.pageSubtitle")}
          extra={
            <Space wrap>
              {activeQuery ? <Tag color="processing">{t("search.queryTag", { query: activeQuery })}</Tag> : null}
              {cached ? <Tag color="success">{t("search.cachedResult")}</Tag> : null}
            </Space>
          }
        />

        <Space.Compact className="search-bar">
          <Input
            size="large"
            placeholder={t("search.placeholder")}
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            onPressEnter={runSearch}
            prefix={<SearchOutlined />}
          />
          <Button type="primary" size="large" onClick={runSearch} disabled={!userId}>
            {t("search.searchBtn")}
          </Button>
        </Space.Compact>
      </Card>

      {!userId ? (
        <Card className="surface-card">
          <Empty description={t("search.emptyHint")} />
        </Card>
      ) : (
        <>
          <div className="stats-grid search-stats">
            <Card className="surface-card">
              <Statistic title={t("search.statMatches")} value={items.length} />
            </Card>
            <Card className="surface-card">
              <Statistic title={t("search.statTopScore")} value={items[0]?.final_score ?? 0} precision={2} />
            </Card>
          </div>

          <Card className="surface-card">
            {activeQuery ? (
              <List
                loading={loading}
                dataSource={items}
                locale={{ emptyText: t("search.emptyMatch") }}
                renderItem={(item, index) => (
                  <List.Item className="search-item">
                    <div className="search-item-body">
                      <Space wrap>
                        <Tag color="blue">{t(categoryKeys[item.memory.category])}</Tag>
                        <Tag color="gold">{t("search.importanceTag", { value: item.memory.importance })}</Tag>
                        <Tag color="geekblue">{t("search.rankTag", { index: index + 1 })}</Tag>
                      </Space>

                      <Typography.Title level={4} className="search-item-title">
                        <HighlightText
                          text={item.memory.content}
                          terms={item.matched_terms.length > 0 ? item.matched_terms : [activeQuery]}
                        />
                      </Typography.Title>

                      <Space wrap className="search-score-row">
                        <Tag>{t("search.finalTag", { value: item.final_score.toFixed(2) })}</Tag>
                        <Tag>{t("search.relevanceTag", { value: item.relevance_score.toFixed(2) })}</Tag>
                        <Tag>{t("search.recencyTag", { value: item.recency_score.toFixed(2) })}</Tag>
                        <Tag>{formatDateTime(item.memory.updated_at)}</Tag>
                      </Space>
                    </div>
                  </List.Item>
                )}
              />
            ) : (
              <Empty description={t("search.emptyStart")} />
            )}
          </Card>
        </>
      )}
    </div>
  );
}
