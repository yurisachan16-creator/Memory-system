import { Button, Card, Typography } from "antd";
import { useNavigate } from "react-router-dom";

export function NotFoundPage() {
  const navigate = useNavigate();

  return (
    <Card className="surface-card centered-panel">
      <Typography.Title level={2}>Nothing lives here</Typography.Title>
      <Typography.Paragraph>
        The requested page does not exist inside the memory console.
      </Typography.Paragraph>
      <Button type="primary" onClick={() => navigate("/memories")}>
        Back to memories
      </Button>
    </Card>
  );
}
