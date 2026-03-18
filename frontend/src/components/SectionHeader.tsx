import { Space, Typography } from "antd";
import type { ReactNode } from "react";

interface SectionHeaderProps {
  title: string;
  subtitle: string;
  extra?: ReactNode;
}

export function SectionHeader({ title, subtitle, extra }: SectionHeaderProps) {
  return (
    <div className="section-header">
      <Space direction="vertical" size={2}>
        <Typography.Title level={3} className="section-title">
          {title}
        </Typography.Title>
        <Typography.Paragraph className="section-subtitle">{subtitle}</Typography.Paragraph>
      </Space>
      {extra}
    </div>
  );
}
