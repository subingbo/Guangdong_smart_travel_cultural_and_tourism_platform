import { useCallback, useEffect, useState } from 'react';
import {
  Button,
  Card,
  Empty,
  Form,
  Input,
  InputNumber,
  Modal,
  Popconfirm,
  Select,
  Space,
  Switch,
  Table,
  Tabs,
  Tag,
  message,
} from 'antd';
import { DeleteOutlined, EditOutlined, PlusOutlined, ReloadOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { api, Knowledge, Order, Page, Product } from '../api/client';
import { useAuth } from '../context/AuthContext';
import SafeImage from '../components/SafeImage';

const COPY = {
  zh: {
    title: '运营管理后台',
    subtitle: '维护知识库、商品目录和订单。所有写操作仅管理员可用。',
    knowledge: '知识库',
    products: '商品',
    orders: '订单',
    newKnowledge: '新增知识',
    newProduct: '新增商品',
    edit: '编辑',
    delete: '下架',
    deleteConfirm: '确定下架这个商品？',
    deleteKnowledgeConfirm: '确定删除这条知识？',
    save: '保存',
    cancel: '取消',
    refresh: '刷新',
    noAccess: '当前账号不是管理员，无法访问后台。',
    loginRequired: '请先登录管理员账号。',
    empty: '暂无数据',
    titleCn: '中文标题',
    titleEn: '英文标题',
    contentCn: '中文内容',
    contentEn: '英文内容',
    tags: '标签',
    nameCn: '中文名称',
    nameEn: '英文名称',
    descCn: '中文描述',
    descEn: '英文描述',
    price: '价格（元）',
    sold: '已售',
    rating: '评分',
    category: '分类',
    imageUrl: '图片 URL',
    stock: '库存',
    active: '上架',
    status: '状态',
    orderNo: '订单号',
    total: '金额',
    createdAt: '创建时间',
    items: '商品',
    pending: '待结算',
    paid: '已支付',
    completed: '已完成',
    canceled: '已取消',
    all: '全部',
    ticket: '门票',
    snack: '美食',
    souvenir: '文创',
    experience: '体验',
    coupon: '优惠券',
    saved: '已保存',
    deleted: '已删除',
    statusUpdated: '状态已更新',
    productsCount: '件商品',
  },
  en: {
    title: 'Operations Console',
    subtitle: 'Maintain the knowledge base, product catalogue and orders. Writes are admin-only.',
    knowledge: 'Knowledge',
    products: 'Products',
    orders: 'Orders',
    newKnowledge: 'New entry',
    newProduct: 'New product',
    edit: 'Edit',
    delete: 'Disable',
    deleteConfirm: 'Disable this product?',
    deleteKnowledgeConfirm: 'Delete this knowledge entry?',
    save: 'Save',
    cancel: 'Cancel',
    refresh: 'Refresh',
    noAccess: 'This account is not an administrator.',
    loginRequired: 'Sign in with an admin account first.',
    empty: 'No data',
    titleCn: 'Chinese title',
    titleEn: 'English title',
    contentCn: 'Chinese content',
    contentEn: 'English content',
    tags: 'Tags',
    nameCn: 'Chinese name',
    nameEn: 'English name',
    descCn: 'Chinese description',
    descEn: 'English description',
    price: 'Price (CNY)',
    sold: 'Sold',
    rating: 'Rating',
    category: 'Category',
    imageUrl: 'Image URL',
    stock: 'Stock',
    active: 'Active',
    status: 'Status',
    orderNo: 'Order',
    total: 'Total',
    createdAt: 'Created',
    items: 'Items',
    pending: 'Pending',
    paid: 'Paid',
    completed: 'Completed',
    canceled: 'Canceled',
    all: 'All',
    ticket: 'Tickets',
    snack: 'Food',
    souvenir: 'Crafts',
    experience: 'Experiences',
    coupon: 'Coupons',
    saved: 'Saved',
    deleted: 'Deleted',
    statusUpdated: 'Status updated',
    productsCount: 'items',
  },
} as const;

type ProductForm = {
  nameCn: string;
  nameEn: string;
  descriptionCn: string;
  descriptionEn: string;
  priceYuan: number;
  category: string;
  imageUrl: string;
  stock: number;
  active: boolean;
};

type KnowledgeForm = {
  titleCn: string;
  titleEn: string;
  contentCn: string;
  contentEn: string;
  tags: string[];
};

function money(cents: number) {
  return `¥${(cents / 100).toFixed(cents % 100 === 0 ? 0 : 2)}`;
}

export default function AdminPage() {
  const { i18n } = useTranslation();
  const { user, openAuth } = useAuth();
  const L = i18n.language === 'zh' ? COPY.zh : COPY.en;

  const [tab, setTab] = useState('knowledge');
  const [knowledge, setKnowledge] = useState<Knowledge[]>([]);
  const [products, setProducts] = useState<Product[]>([]);
  const [orders, setOrders] = useState<Order[]>([]);
  const [loading, setLoading] = useState(false);
  const [knowledgeOpen, setKnowledgeOpen] = useState(false);
  const [productOpen, setProductOpen] = useState(false);
  const [editingKnowledge, setEditingKnowledge] = useState<Knowledge | null>(null);
  const [editingProduct, setEditingProduct] = useState<Product | null>(null);
  const [saving, setSaving] = useState(false);
  const [knowledgeForm] = Form.useForm<KnowledgeForm>();
  const [productForm] = Form.useForm<ProductForm>();

  const isAdmin = user?.role === 'admin';

  const load = useCallback(async () => {
    if (!isAdmin) return;
    setLoading(true);
    try {
      if (tab === 'knowledge') {
        const res = await api.get<{ items: Knowledge[] }>('/api/admin/knowledge?page=1&pageSize=50');
        setKnowledge(res.items || []);
      } else if (tab === 'products') {
        const res = await api.get<Page<Product>>('/api/admin/products?page=1&pageSize=50');
        setProducts(res.items || []);
      } else {
        const res = await api.get<Page<Order>>('/api/admin/orders?page=1&pageSize=50');
        setOrders(res.items || []);
      }
    } catch (e) {
      message.error((e as Error).message);
    } finally {
      setLoading(false);
    }
  }, [tab, isAdmin]);

  useEffect(() => {
    void load();
  }, [load]);

  if (!user) {
    return (
      <Card className="admin-gate">
        <Empty description={L.loginRequired} />
        <Button type="primary" onClick={openAuth}>
          {i18n.language === 'zh' ? '登录' : 'Sign in'}
        </Button>
      </Card>
    );
  }

  if (!isAdmin) {
    return (
      <Card className="admin-gate">
        <Empty description={L.noAccess} />
      </Card>
    );
  }

  const openKnowledge = (item?: Knowledge) => {
    setEditingKnowledge(item || null);
    knowledgeForm.setFieldsValue(
      item
        ? {
            titleCn: item.titleCn,
            titleEn: item.titleEn,
            contentCn: item.contentCn,
            contentEn: item.contentEn,
            tags: item.tags || [],
          }
        : { titleCn: '', titleEn: '', contentCn: '', contentEn: '', tags: [] },
    );
    setKnowledgeOpen(true);
  };

  const saveKnowledge = async () => {
    const values = await knowledgeForm.validateFields();
    setSaving(true);
    try {
      if (editingKnowledge) {
        await api.put(`/api/admin/knowledge/${editingKnowledge.id}`, values);
      } else {
        await api.post('/api/admin/knowledge', values);
      }
      message.success(L.saved);
      setKnowledgeOpen(false);
      void load();
    } catch (e) {
      message.error((e as Error).message);
    } finally {
      setSaving(false);
    }
  };

  const deleteKnowledge = async (id: number) => {
    try {
      await api.del(`/api/admin/knowledge/${id}`);
      message.success(L.deleted);
      void load();
    } catch (e) {
      message.error((e as Error).message);
    }
  };

  const openProduct = (item?: Product) => {
    setEditingProduct(item || null);
    productForm.setFieldsValue(
      item
        ? {
            nameCn: item.nameCn,
            nameEn: item.nameEn,
            descriptionCn: item.descriptionCn,
            descriptionEn: item.descriptionEn,
            priceYuan: item.priceCents / 100,
            category: item.category,
            imageUrl: item.imageUrl,
            stock: item.stock,
            active: item.active,
          }
        : {
            nameCn: '',
            nameEn: '',
            descriptionCn: '',
            descriptionEn: '',
            priceYuan: 0,
            category: 'souvenir',
            imageUrl: '',
            stock: 0,
            active: true,
          },
    );
    setProductOpen(true);
  };

  const saveProduct = async () => {
    const values = await productForm.validateFields();
    const payload = {
      nameCn: values.nameCn,
      nameEn: values.nameEn,
      descriptionCn: values.descriptionCn,
      descriptionEn: values.descriptionEn,
      priceCents: Math.round((values.priceYuan || 0) * 100),
      category: values.category,
      imageUrl: values.imageUrl,
      stock: values.stock || 0,
      active: values.active,
    };
    setSaving(true);
    try {
      if (editingProduct) {
        await api.put(`/api/admin/products/${editingProduct.id}`, payload);
      } else {
        await api.post('/api/admin/products', payload);
      }
      message.success(L.saved);
      setProductOpen(false);
      void load();
    } catch (e) {
      message.error((e as Error).message);
    } finally {
      setSaving(false);
    }
  };

  const disableProduct = async (id: number) => {
    try {
      await api.del(`/api/admin/products/${id}`);
      message.success(L.deleted);
      void load();
    } catch (e) {
      message.error((e as Error).message);
    }
  };

  const updateOrderStatus = async (id: number, status: string) => {
    try {
      await api.put(`/api/admin/orders/${id}/status`, { status });
      message.success(L.statusUpdated);
      void load();
    } catch (e) {
      message.error((e as Error).message);
    }
  };

  const categoryOptions = [
    { value: 'ticket', label: L.ticket },
    { value: 'snack', label: L.snack },
    { value: 'souvenir', label: L.souvenir },
    { value: 'experience', label: L.experience },
    { value: 'coupon', label: L.coupon },
  ];

  const statusOptions = [
    { value: 'pending', label: L.pending },
    { value: 'paid', label: L.paid },
    { value: 'completed', label: L.completed },
    { value: 'canceled', label: L.canceled },
  ];

  return (
    <div className="admin-page">
      <div className="admin-hero">
        <div>
          <h1>{L.title}</h1>
          <p>{L.subtitle}</p>
        </div>
        <Button icon={<ReloadOutlined />} onClick={() => void load()} loading={loading}>
          {L.refresh}
        </Button>
      </div>

      <Tabs
        activeKey={tab}
        onChange={setTab}
        items={[
          {
            key: 'knowledge',
            label: `${L.knowledge} (${knowledge.length || ''})`,
            children: (
              <Card
                title={L.knowledge}
                extra={
                  <Button type="primary" icon={<PlusOutlined />} onClick={() => openKnowledge()}>
                    {L.newKnowledge}
                  </Button>
                }
              >
                <Table
                  rowKey="id"
                  loading={loading}
                  dataSource={knowledge}
                  pagination={false}
                  locale={{ emptyText: L.empty }}
                  columns={[
                    { title: L.titleCn, dataIndex: 'titleCn' },
                    { title: L.titleEn, dataIndex: 'titleEn' },
                    {
                      title: L.tags,
                      dataIndex: 'tags',
                      render: (tags: string[]) => (tags || []).map((t) => <Tag key={t}>{t}</Tag>),
                    },
                    {
                      title: '',
                      width: 160,
                      render: (_, row) => (
                        <Space>
                          <Button size="small" icon={<EditOutlined />} onClick={() => openKnowledge(row)}>
                            {L.edit}
                          </Button>
                          <Popconfirm
                            title={L.deleteKnowledgeConfirm}
                            onConfirm={() => void deleteKnowledge(row.id)}
                          >
                            <Button size="small" danger icon={<DeleteOutlined />} />
                          </Popconfirm>
                        </Space>
                      ),
                    },
                  ]}
                />
              </Card>
            ),
          },
          {
            key: 'products',
            label: L.products,
            children: (
              <Card
                title={L.products}
                extra={
                  <Button type="primary" icon={<PlusOutlined />} onClick={() => openProduct()}>
                    {L.newProduct}
                  </Button>
                }
              >
                <Table
                  rowKey="id"
                  loading={loading}
                  dataSource={products}
                  pagination={false}
                  locale={{ emptyText: L.empty }}
                  columns={[
                    {
                      title: '',
                      width: 68,
                      render: (_, row) => (
                        <SafeImage className="admin-thumb" src={row.imageUrl} alt={row.nameEn} />
                      ),
                    },
                    { title: L.nameCn, dataIndex: 'nameCn' },
                    {
                      title: L.category,
                      dataIndex: 'category',
                      render: (v: string) => categoryOptions.find((o) => o.value === v)?.label || v,
                    },
                    {
                      title: L.price,
                      dataIndex: 'priceCents',
                      render: (v: number) => money(v),
                    },
                    { title: L.sold, dataIndex: 'soldCount' },
                    {
                      title: L.rating,
                      dataIndex: 'rating',
                      render: (v: number) => (v ? v.toFixed(1) : '—'),
                    },
                    { title: L.stock, dataIndex: 'stock' },
                    {
                      title: L.active,
                      dataIndex: 'active',
                      render: (v: boolean) => (
                        <Tag color={v ? 'green' : 'default'}>
                          {v ? (i18n.language === 'zh' ? '上架' : 'Active') : i18n.language === 'zh' ? '下架' : 'Hidden'}
                        </Tag>
                      ),
                    },
                    {
                      title: '',
                      width: 170,
                      render: (_, row) => (
                        <Space>
                          <Button size="small" icon={<EditOutlined />} onClick={() => openProduct(row)}>
                            {L.edit}
                          </Button>
                          <Popconfirm title={L.deleteConfirm} onConfirm={() => void disableProduct(row.id)}>
                            <Button size="small" danger icon={<DeleteOutlined />} />
                          </Popconfirm>
                        </Space>
                      ),
                    },
                  ]}
                />
              </Card>
            ),
          },
          {
            key: 'orders',
            label: L.orders,
            children: (
              <Card title={L.orders}>
                <Table
                  rowKey="id"
                  loading={loading}
                  dataSource={orders}
                  pagination={false}
                  locale={{ emptyText: L.empty }}
                  columns={[
                    { title: L.orderNo, dataIndex: 'orderNo' },
                    {
                      title: L.items,
                      dataIndex: 'items',
                      render: (items: Order['items']) =>
                        `${(items || []).reduce((n, item) => n + item.quantity, 0)} ${L.productsCount}`,
                    },
                    {
                      title: L.total,
                      dataIndex: 'totalCents',
                      render: (v: number) => money(v),
                    },
                    {
                      title: L.createdAt,
                      dataIndex: 'createdAt',
                      render: (v: string) => new Date(v).toLocaleString(i18n.language === 'zh' ? 'zh-CN' : 'en-US'),
                    },
                    {
                      title: L.status,
                      dataIndex: 'status',
                      width: 180,
                      render: (status: string, row) => (
                        <Select
                          size="small"
                          value={status}
                          style={{ width: 130 }}
                          options={statusOptions}
                          onChange={(value) => void updateOrderStatus(row.id, value)}
                        />
                      ),
                    },
                  ]}
                />
              </Card>
            ),
          },
        ]}
      />

      <Modal
        open={knowledgeOpen}
        title={editingKnowledge ? L.edit : L.newKnowledge}
        onCancel={() => setKnowledgeOpen(false)}
        onOk={() => void saveKnowledge()}
        okText={L.save}
        cancelText={L.cancel}
        confirmLoading={saving}
        width={720}
        destroyOnHidden
      >
        <Form form={knowledgeForm} layout="vertical">
          <Space size="middle" style={{ width: '100%' }} align="start">
            <Form.Item name="titleCn" label={L.titleCn} rules={[{ required: true }]} style={{ flex: 1 }}>
              <Input />
            </Form.Item>
            <Form.Item name="titleEn" label={L.titleEn} rules={[{ required: true }]} style={{ flex: 1 }}>
              <Input />
            </Form.Item>
          </Space>
          <Form.Item name="contentCn" label={L.contentCn}>
            <Input.TextArea rows={4} />
          </Form.Item>
          <Form.Item name="contentEn" label={L.contentEn}>
            <Input.TextArea rows={4} />
          </Form.Item>
          <Form.Item name="tags" label={L.tags}>
            <Select mode="tags" tokenSeparators={[',']} />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        open={productOpen}
        title={editingProduct ? L.edit : L.newProduct}
        onCancel={() => setProductOpen(false)}
        onOk={() => void saveProduct()}
        okText={L.save}
        cancelText={L.cancel}
        confirmLoading={saving}
        width={760}
        destroyOnHidden
      >
        <Form form={productForm} layout="vertical">
          <Space size="middle" style={{ width: '100%' }} align="start">
            <Form.Item name="nameCn" label={L.nameCn} rules={[{ required: true }]} style={{ flex: 1 }}>
              <Input />
            </Form.Item>
            <Form.Item name="nameEn" label={L.nameEn} rules={[{ required: true }]} style={{ flex: 1 }}>
              <Input />
            </Form.Item>
          </Space>
          <Form.Item name="descriptionCn" label={L.descCn}>
            <Input.TextArea rows={2} />
          </Form.Item>
          <Form.Item name="descriptionEn" label={L.descEn}>
            <Input.TextArea rows={2} />
          </Form.Item>
          <Space size="middle" style={{ width: '100%' }} align="start">
            <Form.Item name="priceYuan" label={L.price} rules={[{ required: true }]} style={{ flex: 1 }}>
              <InputNumber min={0} precision={2} style={{ width: '100%' }} />
            </Form.Item>
            <Form.Item name="stock" label={L.stock} rules={[{ required: true }]} style={{ flex: 1 }}>
              <InputNumber min={0} style={{ width: '100%' }} />
            </Form.Item>
            <Form.Item name="category" label={L.category} style={{ flex: 1 }}>
              <Select options={categoryOptions} />
            </Form.Item>
          </Space>
          <Form.Item name="imageUrl" label={L.imageUrl}>
            <Input placeholder="https://... or /uploads/..." />
          </Form.Item>
          <Form.Item name="active" label={L.active} valuePropName="checked">
            <Switch />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
}
