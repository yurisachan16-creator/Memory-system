# 多语言（i18n）系统设计文档

## 1. 背景与目标

Memory System Console 原为纯英文界面。为满足中英文用户需求，引入轻量多语言系统，支持运行时切换语言，无需刷新页面，语言偏好持久化存储。

**目标语言：**
- `en` — English（默认）
- `zh` — 简体中文

---

## 2. 设计决策

### 2.1 自研轻量方案 vs 第三方库

| 方案 | 优势 | 劣势 |
|------|------|------|
| `react-i18next` | 功能全面（复数、命名空间懒加载等） | 引入重依赖，项目规模过小，学习成本高 |
| **自研 Context + 纯函数**（本方案） | 零额外依赖、类型完全安全、逻辑可控 | 不支持复数形式（当前不需要） |

**结论：** 当前字符串总量约 120 条，两种语言，无复数/性别变体需求，自研方案最合适。

### 2.2 翻译字典结构

采用**命名空间 + 扁平 key** 的两层嵌套结构：

```
namespace.key
```

例如：`"memory.pageTitle"`、`"header.apply"`。

这样做的好处：
- TypeScript 可以通过类型推导自动生成合法 key 的联合类型，拼写错误会在编译时报错。
- 按页面/功能分组，易于维护。

### 2.3 参数插值语法

使用 `{paramName}` 占位符，运行时替换：

```typescript
translate("en", "memory.total", { count: 5 })
// → "5 memories"

translate("zh", "header.activeUser", { id: "alice" })
// → "当前用户：alice"
```

选择 `{param}` 而非 `%s` 或 `{{param}}`：简洁、直观、与主流 i18n 库风格兼容。

### 2.4 持久化策略

语言偏好存储在 `localStorage` 的 `memory-system.language` key 下，与用户 ID 存储的 key 前缀一致，便于统一管理。读取时做防御性处理（`try/catch`），确保在无 localStorage 环境（SSR、测试）下安全降级为英文。

---

## 3. 系统架构

```
src/
├── i18n/
│   ├── index.ts                  # 类型定义 + translate() 函数 + 域键映射
│   └── translations/
│       ├── en-US.ts              # 英文字典（as const，作为类型基准）
│       └── zh-CN.ts              # 中文字典（同构结构）
└── context/
    └── LanguageContext.tsx       # React Context + Provider + useLanguage hook
```

### 3.1 数据流

```
localStorage
    │
    ▼
LanguageProvider
    │  lang state
    ├──────────────────────────────┐
    │  t(key, params?)             │  toggleLang()
    ▼                             ▼
useLanguage() hook           语言切换按钮（AppShell）
    │
    ▼
translate(lang, key, params)
    │
    ▼
translations[lang][ns][key]  →  字符串（替换 {param} 后返回）
```

### 3.2 类型安全

```typescript
// 自动生成所有合法 dot-notation key 的联合类型
type Namespace = keyof typeof enUS;
export type TranslationKey = {
  [N in Namespace]: `${N}.${keyof (typeof enUS)[N] & string}`;
}[Namespace];

// 编译期报错示例
t("memory.typo")   // ❌ TypeScript Error: Argument of type '"memory.typo"' is not assignable
t("memory.edit")   // ✅ OK
```

---

## 4. 切换按钮（UI）

切换按钮位于 **AppShell 顶栏**，紧跟用户 ID 标签之后。

- **英文模式**下，按钮显示 `中文`（即将切换至的目标语言）。
- **中文模式**下，按钮显示 `English`。
- 按钮带有 `TranslationOutlined` 图标，辅以 `title` tooltip 提示双语说明。
- 点击后语言立即切换，无需刷新页面（React state 驱动）。

```tsx
<Button
  icon={<TranslationOutlined />}
  onClick={toggleLang}
  title="Switch language / 切换语言"
>
  {t("header.langToggle")}
</Button>
```

---

## 5. 翻译覆盖范围

| 文件 | 翻译覆盖情况 |
|------|------------|
| `AppShell.tsx` | ✅ 全覆盖（标题、副标题、导航、用户标签、警告） |
| `MemoryPage.tsx` | ✅ 全覆盖（表头、操作按钮、筛选器、Toast 消息） |
| `SearchPage.tsx` | ✅ 全覆盖（标题、搜索框、评分标签、空态文案） |
| `SummaryPage.tsx` | ✅ 全覆盖（标题、统计卡片、Bucket 标题与说明） |
| `MemoryFormModal.tsx` | ✅ 全覆盖（Modal 标题、表单标签、校验消息） |
| 分类/来源标签 | ✅ 通过 `categoryKeys` / `sourceKeys` 动态翻译 |

---

## 6. 扩展指南

### 添加新语言（例如 ja-JP）

1. 在 `src/i18n/translations/` 下创建 `ja-JP.ts`，结构与 `en-US.ts` 完全一致。
2. 在 `src/i18n/index.ts` 中将 `Language` 类型扩展为 `"en" | "zh" | "ja"`，并注册到 `translations` 对象。
3. `LanguageContext` 的 `toggleLang` 逻辑需相应调整（可改为循环切换或下拉菜单选择）。

### 添加新翻译 key

1. 在 `en-US.ts` 和 `zh-CN.ts` 中同一命名空间下添加同名字段。
2. TypeScript 会自动将新 key 纳入 `TranslationKey` 联合类型，可立即在组件中使用 `t("ns.newKey")`。

### 动态语言包加载（未来）

若字典体积增大，可改为动态 `import()` 加载，在 Provider 中用 `useEffect` 异步获取并替换 `translations[lang]`，同时展示加载状态。

---

## 7. 命名空间说明

| 命名空间 | 覆盖内容 |
|---------|--------|
| `common` | 分类标签（preference/identity/goal/context）、来源标签（chat/manual/system） |
| `app` | 应用标题与副标题 |
| `nav` | 顶部导航菜单项 |
| `header` | 用户切换区、语言切换按钮、警告消息 |
| `memory` | 记忆列表页全部 UI 文本 |
| `memoryForm` | 创建/编辑弹窗的标签与校验消息 |
| `search` | 搜索页全部 UI 文本 |
| `summary` | 摘要页全部 UI 文本 |
