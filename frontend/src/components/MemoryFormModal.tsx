import { Form, Input, InputNumber, Modal, Select } from "antd";
import { useEffect } from "react";

import { CATEGORY_OPTIONS, SOURCE_OPTIONS } from "../types/memory";

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
  const createMode = mode === "create";

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
      title={createMode ? "Add memory" : "Edit memory"}
      okText={createMode ? "Create" : "Save"}
      onCancel={onCancel}
      onOk={async () => {
        const values = await form.validateFields();
        await onSubmit(values);
      }}
    >
      <Form form={form} layout="vertical">
        <Form.Item
          label="Content"
          name="content"
          rules={[
            { required: true, message: "Please enter memory content." },
            { min: 2, message: "Content should be at least 2 characters." }
          ]}
        >
          <Input.TextArea rows={5} maxLength={500} showCount />
        </Form.Item>

        <Form.Item label="Category" name="category" rules={[{ required: true, message: "Select a category." }]}>
          <Select options={CATEGORY_OPTIONS} />
        </Form.Item>

        {createMode && (
          <Form.Item label="Source" name="source" rules={[{ required: true, message: "Select a source." }]}>
            <Select options={SOURCE_OPTIONS} />
          </Form.Item>
        )}

        <Form.Item
          label="Importance"
          name="importance"
          rules={[{ required: true, message: "Set importance between 1 and 5." }]}
        >
          <InputNumber min={1} max={5} precision={0} className="full-width" />
        </Form.Item>
      </Form>
    </Modal>
  );
}
