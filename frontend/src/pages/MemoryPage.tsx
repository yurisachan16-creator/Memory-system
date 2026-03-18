import { DeleteOutlined, EditOutlined, PlusOutlined, ReloadOutlined } from "@ant-design/icons";
import { App, Button, Card, Empty, Popconfirm, Select, Space, Table, Tag, Typography } from "antd";
import type { ColumnsType, TablePaginationConfig } from "antd/es/table";
import { useEffect, useMemo, useState } from "react";

import { createMemory, deleteMemory, listMemories, updateMemory } from "../api/memories";
import { MemoryFormModal, type MemoryFormValues } from "../components/MemoryFormModal";
import { SectionHeader } from "../components/SectionHeader";
import { useLanguage } from "../context/LanguageContext";
import { useUser } from "../context/UserContext";
import { categoryKeys } from "../i18n";
import {
  type Category,
  type CreateMemoryPayload,
  type ListMemoriesParams,
  type ListMemoriesResult,
  type Memory,
  type UpdateMemoryPayload,
  formatDateTime
} from "../types/memory";

const emptyResult: ListMemoriesResult = {
  items: [],
  total: 0,
  page: 1,
  page_size: 10
};

export function MemoryPage() {
  const { message } = App.useApp();
  const { userId } = useUser();
  const { t } = useLanguage();
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

  const categoryOptions = useMemo(
    () =>
      (["preference", "identity", "goal", "context"] as Category[]).map((v) => ({
        label: t(categoryKeys[v]),
        value: v
      })),
    [t]
  );

  const sortOptions = useMemo(
    () => [
      { label: t("memory.sortCreated"), value: "created_at" as const },
      { label: t("memory.sortImportance"), value: "importance" as const }
    ],
    [t]
  );

  const orderOptions = useMemo(
    () => [
      { label: t("memory.orderDesc"), value: "desc" as const },
      { label: t("memory.orderAsc"), value: "asc" as const }
    ],
    [t]
  );

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
          void message.error(error instanceof Error ? error.message : t("memory.loadError"));
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
  }, [category, message, order, pagination.current, pagination.pageSize, refreshKey, sortBy, t, userId]);

  const columns: ColumnsType<Memory> = useMemo(
    () => [
      {
        title: t("memory.colContent"),
        dataIndex: "content",
        key: "content",
        render: (value: string) => (
          <Typography.Paragraph ellipsis={{ rows: 2, expandable: true, symbol: "more" }} className="table-content">
            {value}
          </Typography.Paragraph>
        )
      },
      {
        title: t("memory.colCategory"),
        dataIndex: "category",
        key: "category",
        width: 130,
        render: (value: Memory["category"]) => (
          <Tag color="blue">{t(categoryKeys[value])}</Tag>
        )
      },
      {
        title: t("memory.colSource"),
        dataIndex: "source",
        key: "source",
        width: 120
      },
      {
        title: t("memory.colImportance"),
        dataIndex: "importance",
        key: "importance",
        width: 120,
        render: (value: number) => <Tag color="gold">P{value}</Tag>
      },
      {
        title: t("memory.colCreated"),
        dataIndex: "created_at",
        key: "created_at",
        width: 200,
        render: formatDateTime
      },
      {
        title: t("memory.colUpdated"),
        dataIndex: "updated_at",
        key: "updated_at",
        width: 200,
        render: formatDateTime
      },
      {
        title: t("memory.colActions"),
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
              {t("memory.edit")}
            </Button>
            <Popconfirm
              title={t("memory.deleteTitle")}
              description={t("memory.deleteDesc")}
              okText={t("memory.deleteOk")}
              cancelText={t("memory.deleteCancel")}
              onConfirm={async () => {
                try {
                  await deleteMemory(record.id, userId);
                  void message.success(t("memory.deleteSuccess"));
                  const nextTotal = Math.max(result.total - 1, 0);
                  const maxPage = Math.max(Math.ceil(nextTotal / (pagination.pageSize ?? 10)), 1);
                  setPagination((current) => ({
                    ...current,
                    current: Math.min(current.current ?? 1, maxPage)
                  }));
                  setRefreshKey((value) => value + 1);
                } catch (error) {
                  void message.error(error instanceof Error ? error.message : t("memory.deleteError"));
                }
              }}
            >
              <Button size="small" danger icon={<DeleteOutlined />}>
                {t("memory.delete")}
              </Button>
            </Popconfirm>
          </Space>
        )
      }
    ],
    [message, pagination.pageSize, result.total, t, userId]
  );

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
      void message.warning(t("memory.noUserWarning"));
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
        void message.success(t("memory.createSuccess"));
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
        void message.success(t("memory.updateSuccess"));
        refreshCurrentPage();
      }
      setModalOpen(false);
    } catch (error) {
      void message.error(error instanceof Error ? error.message : t("memory.saveError"));
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className="page-stack">
      <Card className="hero-card">
        <SectionHeader
          title={t("memory.pageTitle")}
          subtitle={t("memory.pageSubtitle")}
          extra={
            <Space wrap>
              <Button icon={<ReloadOutlined />} onClick={refreshCurrentPage}>
                {t("memory.refresh")}
              </Button>
              <Button type="primary" icon={<PlusOutlined />} onClick={openCreateModal} disabled={!userId}>
                {t("memory.add")}
              </Button>
            </Space>
          }
        />

        <Space wrap className="filter-bar">
          <Select
            allowClear
            placeholder={t("memory.filterCategory")}
            options={categoryOptions}
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
          <Empty description={t("memory.emptyHint")} />
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
              showTotal: (total) => t("memory.total", { count: total })
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
