import { Form, Input, InputNumber, Modal, Select } from "antd";
import { useEffect, useMemo } from "react";

import { useLanguage } from "../context/LanguageContext";
import { sourceKeys, categoryKeys } from "../i18n";
import type { Category, Source } from "../types/memory";

export interface MemoryFormValues {
  content: string;
  category: string;
  source?: string;
  importance: number;
}

interface MemoryFormModalProps {
  open: boolean;
  mode: "create" | "edit";
  initialValues?: Partial<MemoryFormValues>;
  confirmLoading?: boolean;
  onCancel: () => void;
  onSubmit: (values: MemoryFormValues) => Promise<void>;
}

const defaultValues: MemoryFormValues = {
  content: "",
  category: "context",
  source: "manual",
  importance: 3
};

export function MemoryFormModal({
  open,
  mode,
  initialValues,
  confirmLoading,
  onCancel,
  onSubmit
}: MemoryFormModalProps) {
  const [form] = Form.useForm<MemoryFormValues>();
  const { t } = useLanguage();
  const createMode = mode === "create";

  const categoryOptions = useMemo(
    () =>
      (["preference", "identity", "goal", "context"] as Category[]).map((v) => ({
        label: t(categoryKeys[v]),
        value: v
      })),
    [t]
  );

  const sourceOptions = useMemo(
    () =>
      (["chat", "manual", "system"] as Source[]).map((v) => ({
        label: t(sourceKeys[v]),
        value: v
      })),
    [t]
  );

  useEffect(() => {
    if (!open) {
      form.resetFields();
      return;
    }

    form.setFieldsValue({
      ...defaultValues,
      ...initialValues
    });
  }, [form, initialValues, open]);

  return (
    <Modal
      open={open}
      destroyOnClose
      confirmLoading={confirmLoading}
      title={createMode ? t("memoryForm.titleCreate") : t("memoryForm.titleEdit")}
      okText={createMode ? t("memoryForm.okCreate") : t("memoryForm.okSave")}
      onCancel={onCancel}
      onOk={async () => {
        const values = await form.validateFields();
        await onSubmit(values);
      }}
    >
      <Form form={form} layout="vertical">
        <Form.Item
          label={t("memoryForm.labelContent")}
          name="content"
          rules={[
            { required: true, message: t("memoryForm.ruleContent") },
            { min: 2, message: t("memoryForm.ruleContentMin") }
          ]}
        >
          <Input.TextArea rows={5} maxLength={500} showCount />
        </Form.Item>

        <Form.Item
          label={t("memoryForm.labelCategory")}
          name="category"
          rules={[{ required: true, message: t("memoryForm.ruleCategory") }]}
        >
          <Select options={categoryOptions} />
        </Form.Item>

        {createMode && (
          <Form.Item
            label={t("memoryForm.labelSource")}
            name="source"
            rules={[{ required: true, message: t("memoryForm.ruleSource") }]}
          >
            <Select options={sourceOptions} />
          </Form.Item>
        )}

        <Form.Item
          label={t("memoryForm.labelImportance")}
          name="importance"
          rules={[{ required: true, message: t("memoryForm.ruleImportance") }]}
        >
          <InputNumber min={1} max={5} precision={0} className="full-width" />
        </Form.Item>
      </Form>
    </Modal>
  );
}
