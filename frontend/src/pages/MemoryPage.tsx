import { DeleteOutlined, EditOutlined, PlusOutlined, ReloadOutlined } from "@ant-design/icons";
import { App, Button, Card, Empty, Popconfirm, Select, Space, Table, Tag, Typography } from "antd";
import type { ColumnsType, TablePaginationConfig } from "antd/es/table";
import { useEffect, useState } from "react";

import { createMemory, deleteMemory, listMemories, updateMemory } from "../api/memories";
import { MemoryFormModal, type MemoryFormValues } from "../components/MemoryFormModal";
import { SectionHeader } from "../components/SectionHeader";
import { useUser } from "../context/UserContext";
import {
  CATEGORY_OPTIONS,
  categoryLabelMap,
  formatDateTime,
  type Category,
  type CreateMemoryPayload,
  type ListMemoriesParams,
  type ListMemoriesResult,
  type Memory,
  type UpdateMemoryPayload
} from "../types/memory";

const sortOptions = [
  { label: "Created time", value: "created_at" },
  { label: "Importance", value: "importance" }
];

const orderOptions = [
  { label: "Descending", value: "desc" },
  { label: "Ascending", value: "asc" }
];

const emptyResult: ListMemoriesResult = {
  items: [],
  total: 0,
  page: 1,
  page_size: 10
};

export function MemoryPage() {
  const { message } = App.useApp();
  const { userId } = useUser();
  const [loading, setLoading] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [mode, setMode] = useState<"create" | "edit">("create");
  const [editingMemory, setEditingMemory] = useState<Memory | null>(null);
  const [result, setResult] = useState<ListMemoriesResult>(emptyResult);
  const [category, setCategory] = useState<Category | undefined>();
  const [sortBy, setSortBy] = useState<NonNullable<ListMemoriesParams["sort_by"]>>("created_at");
  const [order, setOrder] = useState<NonNullable<ListMemoriesParams["order"]>>("desc");
  const [refreshKey, setRefreshKey] = useState(0);
  const [pagination, setPagination] = useState<TablePaginationConfig>({
    current: 1,
    pageSize: 10,
    showSizeChanger: true
  });

  useEffect(() => {
    if (!userId) {
      setResult(emptyResult);
      return;
    }

    let cancelled = false;

    async function load() {
      setLoading(true);
      try {
        const data = await listMemories({
          user_id: userId,
          category,
          sort_by: sortBy,
          order,
          page: pagination.current ?? 1,
          page_size: pagination.pageSize ?? 10
        });

        if (!cancelled) {
          setResult(data);
        }
      } catch (error) {
        if (!cancelled) {
          void message.error(error instanceof Error ? error.message : "Failed to load memories");
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
  }, [category, message, order, pagination.current, pagination.pageSize, refreshKey, sortBy, userId]);

  const columns: ColumnsType<Memory> = [
    {
      title: "Content",
      dataIndex: "content",
      key: "content",
      render: (value: string) => (
        <Typography.Paragraph ellipsis={{ rows: 2, expandable: true, symbol: "more" }} className="table-content">
          {value}
        </Typography.Paragraph>
      )
    },
    {
      title: "Category",
      dataIndex: "category",
      key: "category",
      width: 130,
      render: (value: Memory["category"]) => <Tag color="blue">{categoryLabelMap[value]}</Tag>
    },
    {
      title: "Source",
      dataIndex: "source",
      key: "source",
      width: 120
    },
    {
      title: "Importance",
      dataIndex: "importance",
      key: "importance",
      width: 120,
      render: (value: number) => <Tag color="gold">P{value}</Tag>
    },
    {
      title: "Created",
      dataIndex: "created_at",
      key: "created_at",
      width: 200,
      render: formatDateTime
    },
    {
      title: "Updated",
      dataIndex: "updated_at",
      key: "updated_at",
      width: 200,
      render: formatDateTime
    },
    {
      title: "Actions",
      key: "actions",
      width: 140,
      render: (_, record) => (
        <Space>
          <Button
            size="small"
            icon={<EditOutlined />}
            onClick={() => {
              setMode("edit");
              setEditingMemory(record);
              setModalOpen(true);
            }}
          >
            Edit
          </Button>
          <Popconfirm
            title="Delete this memory?"
            description="This will perform a soft delete for the active user."
            okText="Delete"
            cancelText="Cancel"
            onConfirm={async () => {
              try {
                await deleteMemory(record.id, userId);
                void message.success("Memory deleted");
                const nextTotal = Math.max(result.total - 1, 0);
                const maxPage = Math.max(Math.ceil(nextTotal / (pagination.pageSize ?? 10)), 1);
                setPagination((current) => ({
                  ...current,
                  current: Math.min(current.current ?? 1, maxPage)
                }));
                setRefreshKey((value) => value + 1);
              } catch (error) {
                void message.error(error instanceof Error ? error.message : "Failed to delete memory");
              }
            }}
          >
            <Button size="small" danger icon={<DeleteOutlined />}>
              Delete
            </Button>
          </Popconfirm>
        </Space>
      )
    }
  ];

  const refreshCurrentPage = () => {
    setRefreshKey((value) => value + 1);
  };

  const openCreateModal = () => {
    setMode("create");
    setEditingMemory(null);
    setModalOpen(true);
  };

  async function handleSubmit(values: MemoryFormValues) {
    if (!userId) {
      void message.warning("Set a user_id before saving memories");
      return;
    }

    setSubmitting(true);
    try {
      if (mode === "create") {
        const payload: CreateMemoryPayload = {
          user_id: userId,
          content: values.content.trim(),
          category: values.category as CreateMemoryPayload["category"],
          source: (values.source || "manual") as CreateMemoryPayload["source"],
          importance: values.importance
        };
        await createMemory(payload);
        void message.success("Memory created");
        if ((pagination.current ?? 1) === 1) {
          refreshCurrentPage();
        } else {
          setPagination((current) => ({ ...current, current: 1 }));
        }
      } else if (editingMemory) {
        const payload: UpdateMemoryPayload = {
          user_id: userId,
          content: values.content.trim(),
          category: values.category as UpdateMemoryPayload["category"],
          importance: values.importance
        };
        await updateMemory(editingMemory.id, payload);
        void message.success("Memory updated");
        refreshCurrentPage();
      }
      setModalOpen(false);
    } catch (error) {
      void message.error(error instanceof Error ? error.message : "Failed to save memory");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className="page-stack">
      <Card className="hero-card">
        <SectionHeader
          title="Memory Management"
          subtitle="Browse and curate long-term memories with filters, pagination, and controlled edits."
          extra={
            <Space wrap>
              <Button icon={<ReloadOutlined />} onClick={refreshCurrentPage}>
                Refresh
              </Button>
              <Button type="primary" icon={<PlusOutlined />} onClick={openCreateModal} disabled={!userId}>
                Add memory
              </Button>
            </Space>
          }
        />

        <Space wrap className="filter-bar">
          <Select
            allowClear
            placeholder="Filter category"
            options={CATEGORY_OPTIONS}
            value={category}
            onChange={(value) => {
              setCategory(value);
              setPagination((current) => ({ ...current, current: 1 }));
            }}
            className="filter-control"
          />
          <Select
            options={sortOptions}
            value={sortBy}
            onChange={(value) => {
              setSortBy(value);
              setPagination((current) => ({ ...current, current: 1 }));
            }}
            className="filter-control"
          />
          <Select
            options={orderOptions}
            value={order}
            onChange={(value) => {
              setOrder(value);
              setPagination((current) => ({ ...current, current: 1 }));
            }}
            className="filter-control"
          />
        </Space>
      </Card>

      <Card className="surface-card">
        {!userId ? (
          <Empty description="Set a user_id in the top bar to load memories." />
        ) : (
          <Table
            rowKey="id"
            columns={columns}
            dataSource={result.items}
            loading={loading}
            pagination={{
              ...pagination,
              current: result.page,
              pageSize: result.page_size,
              total: result.total,
              showTotal: (total) => `${total} memories`
            }}
            onChange={(nextPagination) => setPagination(nextPagination)}
          />
        )}
      </Card>

      <MemoryFormModal
        open={modalOpen}
        mode={mode}
        initialValues={
          editingMemory
            ? {
                content: editingMemory.content,
                category: editingMemory.category,
                importance: editingMemory.importance
              }
            : undefined
        }
        confirmLoading={submitting}
        onCancel={() => setModalOpen(false)}
        onSubmit={handleSubmit}
      />
    </div>
  );
}
