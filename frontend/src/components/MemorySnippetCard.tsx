import { Card, Space, Tag, Typography } from "antd";

import { categoryLabelMap, formatDateTime, sourceLabelMap, type Memory } from "../types/memory";

interface MemorySnippetCardProps {
  memory: Memory;
}

export function MemorySnippetCard({ memory }: MemorySnippetCardProps) {
  return (
    <Card size="small" className="memory-snippet-card">
      <Space wrap className="memory-meta-row">
        <Tag color="blue">{categoryLabelMap[memory.category]}</Tag>
        <Tag color="volcano">{sourceLabelMap[memory.source]}</Tag>
        <Tag color="gold">Importance {memory.importance}</Tag>
      </Space>

      <Typography.Paragraph className="memory-content" ellipsis={{ rows: 3 }}>
        {memory.content}
      </Typography.Paragraph>

      <Typography.Text type="secondary">Updated {formatDateTime(memory.updated_at)}</Typography.Text>
    </Card>
  );
}
