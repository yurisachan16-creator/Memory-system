import { SearchOutlined } from "@ant-design/icons";
import { App, Button, Card, Empty, Input, List, Space, Statistic, Tag, Typography } from "antd";
import { useEffect, useState } from "react";

import { searchMemories } from "../api/memories";
import { HighlightText } from "../components/HighlightText";
import { SectionHeader } from "../components/SectionHeader";
import { useUser } from "../context/UserContext";
import { categoryLabelMap, formatDateTime, type SearchItem } from "../types/memory";

export function SearchPage() {
  const { message } = App.useApp();
  const { userId } = useUser();
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
          void message.error(error instanceof Error ? error.message : "Search failed");
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
  }, [activeQuery, message, userId]);

  const runSearch = () => {
    const nextQuery = query.trim();
    if (!userId) {
      void message.warning("Set a user_id before searching");
      return;
    }
    if (!nextQuery) {
      void message.warning("Enter a search query");
      return;
    }
    setActiveQuery(nextQuery);
  };

  return (
    <div className="page-stack">
      <Card className="hero-card">
        <SectionHeader
          title="Memory Search"
          subtitle="Query the active user's memory bank and inspect relevance, recency, and matched terms."
          extra={
            <Space wrap>
              {activeQuery ? <Tag color="processing">Query: {activeQuery}</Tag> : null}
              {cached ? <Tag color="success">Cached result</Tag> : null}
            </Space>
          }
        />

        <Space.Compact className="search-bar">
          <Input
            size="large"
            placeholder="Search for a preference, goal, or context clue"
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            onPressEnter={runSearch}
            prefix={<SearchOutlined />}
          />
          <Button type="primary" size="large" onClick={runSearch} disabled={!userId}>
            Search
          </Button>
        </Space.Compact>
      </Card>

      {!userId ? (
        <Card className="surface-card">
          <Empty description="Set a user_id in the top bar to search memories." />
        </Card>
      ) : (
        <>
          <div className="stats-grid search-stats">
            <Card className="surface-card">
              <Statistic title="Matches" value={items.length} />
            </Card>
            <Card className="surface-card">
              <Statistic title="Top score" value={items[0]?.final_score ?? 0} precision={2} />
            </Card>
          </div>

          <Card className="surface-card">
            {activeQuery ? (
              <List
                loading={loading}
                dataSource={items}
                locale={{ emptyText: "No matching memories found." }}
                renderItem={(item, index) => (
                  <List.Item className="search-item">
                    <div className="search-item-body">
                      <Space wrap>
                        <Tag color="blue">{categoryLabelMap[item.memory.category]}</Tag>
                        <Tag color="gold">Importance {item.memory.importance}</Tag>
                        <Tag color="geekblue">Rank #{index + 1}</Tag>
                      </Space>

                      <Typography.Title level={4} className="search-item-title">
                        <HighlightText
                          text={item.memory.content}
                          terms={item.matched_terms.length > 0 ? item.matched_terms : [activeQuery]}
                        />
                      </Typography.Title>

                      <Space wrap className="search-score-row">
                        <Tag>Final {item.final_score.toFixed(2)}</Tag>
                        <Tag>Relevance {item.relevance_score.toFixed(2)}</Tag>
                        <Tag>Recency {item.recency_score.toFixed(2)}</Tag>
                        <Tag>{formatDateTime(item.memory.updated_at)}</Tag>
                      </Space>
                    </div>
                  </List.Item>
                )}
              />
            ) : (
              <Empty description="Run a search to see the most relevant memories." />
            )}
          </Card>
        </>
      )}
    </div>
  );
}
