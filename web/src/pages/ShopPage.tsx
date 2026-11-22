import { useCallback, useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Avatar,
  Badge,
  Button,
  Card,
  Drawer,
  Empty,
  Input,
  Modal,
  Rate,
  Spin,
  Tag,
  message,
} from 'antd';
import {
  DeleteOutlined,
  MinusOutlined,
  PlusOutlined,
  SearchOutlined,
  ShoppingCartOutlined,
  ShopOutlined,
  ThunderboltOutlined,
} from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { api, assetUrl, CartView, Order, Page, Product, ProductReview } from '../api/client';
import { useAuth } from '../context/AuthContext';
import SafeImage from '../components/SafeImage';

const COPY = {
  zh: {
    eyebrow: '南粤风物',
    title: '把岭南味道带回家',
    subtitle: '门票、早茶、文创与旅拍体验，一站选购。线上下单，到店出示订单号结算。',
    search: '搜索商品、门票或体验',
    searchAction: '搜索',
    all: '全部',
    ticket: '门票',
    snack: '美食',
    souvenir: '文创',
    experience: '体验',
    coupon: '优惠券',
    cart: '购物车',
    orders: '我的订单',
    add: '加入购物车',
    out: '暂时售罄',
    stock: '库存',
    empty: '没有找到商品，换个关键词试试。',
    cartEmpty: '购物车还是空的，去逛逛吧。',
    total: '合计',
    checkout: '提交订单',
    checkoutHint: '提交后生成预订单，到店出示订单号结算。',
    successTitle: '下单成功',
    successBody: '预订单已生成，到店出示订单号即可。',
    viewOrders: '查看订单',
    continue: '继续逛逛',
    loginFirst: '请先登录再下单。',
    added: '已加入购物车',
    detail: '商品详情',
    close: '关闭',
    remove: '删除',
    free: '免费',
    qty: '数量',
    sold: '已售',
    reviews: '用户评价',
    noReviews: '还没有评价，欢迎成为第一位分享体验的人。',
    pending: '待结算',
    paid: '已支付',
    completed: '已完成',
    canceled: '已取消',
  },
  en: {
    eyebrow: 'Lingnan Market',
    title: 'Bring a taste of Guangdong home',
    subtitle: 'Tickets, dim sum, crafts and photo walks in one place. Order online and settle at the venue.',
    search: 'Search products, tickets or experiences',
    searchAction: 'Search',
    all: 'All',
    ticket: 'Tickets',
    snack: 'Food',
    souvenir: 'Crafts',
    experience: 'Experiences',
    coupon: 'Coupons',
    cart: 'Cart',
    orders: 'My Orders',
    add: 'Add to cart',
    out: 'Sold out',
    stock: 'Stock',
    empty: 'No products found. Try another keyword.',
    cartEmpty: 'Your cart is empty. Go explore the market.',
    total: 'Total',
    checkout: 'Place order',
    checkoutHint: 'Creates a reservation; show your order number at the venue to settle.',
    successTitle: 'Order created',
    successBody: 'Your reservation is ready. Show the order number at the venue.',
    viewOrders: 'View orders',
    continue: 'Keep shopping',
    loginFirst: 'Please sign in before checking out.',
    added: 'Added to cart',
    detail: 'Product details',
    close: 'Close',
    remove: 'Remove',
    free: 'Free',
    qty: 'Qty',
    sold: 'sold',
    reviews: 'Reviews',
    noReviews: 'No reviews yet. Be the first to share your experience.',
    pending: 'Pending',
    paid: 'Paid',
    completed: 'Completed',
    canceled: 'Canceled',
  },
} as const;

const CATEGORY_KEYS = ['all', 'ticket', 'snack', 'souvenir', 'experience', 'coupon'] as const;

function money(cents: number, freeLabel: string) {
  if (cents <= 0) return freeLabel;
  return `¥${(cents / 100).toFixed(cents % 100 === 0 ? 0 : 2)}`;
}

export default function ShopPage() {
  const { i18n } = useTranslation();
  const { user, openAuth } = useAuth();
  const navigate = useNavigate();
  const L = i18n.language === 'zh' ? COPY.zh : COPY.en;

  const [products, setProducts] = useState<Product[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [category, setCategory] = useState<(typeof CATEGORY_KEYS)[number]>('all');
  const [q, setQ] = useState('');
  const [loading, setLoading] = useState(true);
  const [cart, setCart] = useState<CartView>({ items: [], totalCents: 0, count: 0 });
  const [cartOpen, setCartOpen] = useState(false);
  const [detail, setDetail] = useState<Product | null>(null);
  const [checkingOut, setCheckingOut] = useState(false);
  const [lastOrder, setLastOrder] = useState<Order | null>(null);
  const [successOpen, setSuccessOpen] = useState(false);
  const [addingId, setAddingId] = useState<number | null>(null);
  const [reviews, setReviews] = useState<ProductReview[]>([]);
  const [reviewsLoading, setReviewsLoading] = useState(false);

  const loadProducts = useCallback(async () => {
    setLoading(true);
    try {
      const params = new URLSearchParams({
        page: String(page),
        pageSize: '12',
        category,
        q,
      });
      const res = await api.get<Page<Product>>(`/api/shop/products?${params.toString()}`);
      setProducts(res.items || []);
      setTotal(res.total || 0);
    } catch {
      setProducts([]);
      setTotal(0);
    } finally {
      setLoading(false);
    }
  }, [page, category, q]);

  const loadCart = useCallback(async () => {
    if (!user) {
      setCart({ items: [], totalCents: 0, count: 0 });
      return;
    }
    try {
      setCart(await api.get<CartView>('/api/cart'));
    } catch {
      setCart({ items: [], totalCents: 0, count: 0 });
    }
  }, [user?.id]);

  useEffect(() => {
    void loadProducts();
  }, [loadProducts]);

  useEffect(() => {
    void loadCart();
  }, [loadCart]);

  useEffect(() => {
    if (!detail) {
      setReviews([]);
      return;
    }
    setReviewsLoading(true);
    api
      .get<ProductReview[]>(`/api/shop/products/${detail.id}/reviews`)
      .then((data) => setReviews(data || []))
      .catch(() => setReviews([]))
      .finally(() => setReviewsLoading(false));
  }, [detail?.id]);

  const addToCart = async (product: Product, quantity = 1) => {
    if (!user) {
      openAuth();
      return;
    }
    setAddingId(product.id);
    try {
      setCart(await api.post<CartView>('/api/cart', { productId: product.id, quantity }));
      message.success(L.added);
    } catch (e) {
      message.error((e as Error).message);
    } finally {
      setAddingId(null);
    }
  };

  const updateQuantity = async (productId: number, quantity: number) => {
    try {
      setCart(await api.put<CartView>(`/api/cart/${productId}`, { quantity }));
    } catch (e) {
      message.error((e as Error).message);
    }
  };

  const removeItem = async (productId: number) => {
    try {
      setCart(await api.del<CartView>(`/api/cart/${productId}`));
    } catch (e) {
      message.error((e as Error).message);
    }
  };

  const checkout = async () => {
    if (!user) {
      openAuth();
      return;
    }
    if (cart.items.length === 0) {
      message.warning(L.cartEmpty);
      return;
    }
    setCheckingOut(true);
    try {
      const idem =
        typeof crypto !== 'undefined' && 'randomUUID' in crypto
          ? crypto.randomUUID()
          : `${Date.now()}-${Math.random().toString(16).slice(2)}`;
      const order = await api.post<Order>(
        '/api/shop/orders',
        undefined,
        { 'Idempotency-Key': idem },
      );
      setLastOrder(order);
      setCart({ items: [], totalCents: 0, count: 0 });
      setCartOpen(false);
      setSuccessOpen(true);
    } catch (e) {
      message.error((e as Error).message);
    } finally {
      setCheckingOut(false);
    }
  };

  const categoryLabel = (key: (typeof CATEGORY_KEYS)[number]) => {
    switch (key) {
      case 'ticket':
        return L.ticket;
      case 'snack':
        return L.snack;
      case 'souvenir':
        return L.souvenir;
      case 'experience':
        return L.experience;
      case 'coupon':
        return L.coupon;
      default:
        return L.all;
    }
  };

  const productName = (p: Product) => (i18n.language === 'zh' ? p.nameCn : p.nameEn);
  const productDesc = (p: Product) => (i18n.language === 'zh' ? p.descriptionCn : p.descriptionEn);

  const statusLabel = (status: Order['status']) => {
    switch (status) {
      case 'paid':
        return L.paid;
      case 'completed':
        return L.completed;
      case 'canceled':
        return L.canceled;
      default:
        return L.pending;
    }
  };

  return (
    <div className="shop-page">
      <div className="shop-hero">
        <div className="shop-hero-copy">
          <div className="shop-eyebrow">
            <ShopOutlined /> {L.eyebrow}
          </div>
          <h1>{L.title}</h1>
          <p>{L.subtitle}</p>
          <Input
            className="shop-search"
            size="large"
            allowClear
            value={q}
            onChange={(e) => setQ(e.target.value)}
            onPressEnter={() => {
              setPage(1);
              void loadProducts();
            }}
            prefix={<SearchOutlined />}
            placeholder={L.search}
            suffix={
              <Button
                type="primary"
                size="small"
                onClick={() => {
                  setPage(1);
                  void loadProducts();
                }}
              >
                {L.searchAction}
              </Button>
            }
          />
        </div>
        <div className="shop-hero-actions">
          <Button icon={<ThunderboltOutlined />} onClick={() => navigate('/orders')}>
            {L.orders}
          </Button>
          <Badge count={cart.count} size="small">
            <Button type="primary" icon={<ShoppingCartOutlined />} onClick={() => setCartOpen(true)}>
              {L.cart}
            </Button>
          </Badge>
        </div>
      </div>

      <div className="shop-categories">
        {CATEGORY_KEYS.map((key) => (
          <button
            key={key}
            type="button"
            className={`shop-category${category === key ? ' active' : ''}`}
            onClick={() => {
              setCategory(key);
              setPage(1);
            }}
          >
            {categoryLabel(key)}
          </button>
        ))}
      </div>

      {loading ? (
        <div className="shop-loading">
          <Spin />
        </div>
      ) : products.length === 0 ? (
        <Empty description={L.empty} />
      ) : (
        <div className="shop-grid">
          {products.map((p) => (
            <article key={p.id} className="shop-card" onClick={() => setDetail(p)}>
              <div className="shop-card-image">
                <SafeImage src={p.imageUrl} alt={productName(p)} />
                <Tag className="shop-category-tag">{categoryLabel(p.category as never)}</Tag>
              </div>
              <div className="shop-card-body">
                <h3>{productName(p)}</h3>
                <p>{productDesc(p)}</p>
                <div className="shop-card-rating">
                  <Rate disabled allowHalf value={p.rating} />
                  <span>
                    {p.rating.toFixed(1)} · {L.sold} {p.soldCount.toLocaleString()}
                  </span>
                </div>
                <div className="shop-card-footer">
                  <div>
                    <b className="shop-price">{money(p.priceCents, L.free)}</b>
                    <small className="shop-stock">
                      {L.stock} {p.stock}
                    </small>
                  </div>
                  <Button
                    type="primary"
                    size="small"
                    icon={<PlusOutlined />}
                    loading={addingId === p.id}
                    disabled={p.stock <= 0}
                    onClick={(e) => {
                      e.stopPropagation();
                      void addToCart(p);
                    }}
                  >
                    {p.stock > 0 ? L.add : L.out}
                  </Button>
                </div>
              </div>
            </article>
          ))}
        </div>
      )}

      {total > 12 && (
        <div className="shop-pagination">
          <Button disabled={page <= 1} onClick={() => setPage((v) => v - 1)}>
            ‹
          </Button>
          <span>
            {page} / {Math.ceil(total / 12)}
          </span>
          <Button disabled={page >= Math.ceil(total / 12)} onClick={() => setPage((v) => v + 1)}>
            ›
          </Button>
        </div>
      )}

      <Drawer
        title={
          <span>
            <ShoppingCartOutlined /> {L.cart}
          </span>
        }
        open={cartOpen}
        onClose={() => setCartOpen(false)}
        width={420}
        footer={
          <div className="shop-cart-footer">
            <div className="shop-cart-total">
              <span>{L.total}</span>
              <b>{money(cart.totalCents, L.free)}</b>
            </div>
            <Button
              type="primary"
              block
              size="large"
              loading={checkingOut}
              disabled={cart.items.length === 0}
              onClick={() => void checkout()}
            >
              {L.checkout}
            </Button>
            <small className="shop-cart-hint">{L.checkoutHint}</small>
          </div>
        }
      >
        {cart.items.length === 0 ? (
          <Empty description={L.cartEmpty} />
        ) : (
          <div className="shop-cart-list">
            {cart.items.map((item) => (
              <div key={item.productId} className="shop-cart-item">
                <SafeImage
                  className="shop-cart-thumb"
                  src={item.imageUrl}
                  alt={i18n.language === 'zh' ? item.nameCn : item.nameEn}
                />
                <div className="shop-cart-info">
                  <b>{i18n.language === 'zh' ? item.nameCn : item.nameEn}</b>
                  <span>{money(item.priceCents, L.free)}</span>
                  <div className="shop-qty">
                    <Button
                      size="small"
                      icon={<MinusOutlined />}
                      onClick={() => void updateQuantity(item.productId, item.quantity - 1)}
                    />
                    <b>{item.quantity}</b>
                    <Button
                      size="small"
                      icon={<PlusOutlined />}
                      disabled={item.quantity >= item.stock}
                      onClick={() => void updateQuantity(item.productId, item.quantity + 1)}
                    />
                    <Button
                      size="small"
                      type="text"
                      danger
                      icon={<DeleteOutlined />}
                      onClick={() => void removeItem(item.productId)}
                    />
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </Drawer>

      <Modal
        open={!!detail}
        title={L.detail}
        footer={null}
        onCancel={() => setDetail(null)}
        width={760}
      >
        {detail && (
          <div className="shop-detail">
            <SafeImage className="shop-detail-image" src={detail.imageUrl} alt={productName(detail)} />
            <div className="shop-detail-body">
              <Tag>{categoryLabel(detail.category as never)}</Tag>
              <h2>{productName(detail)}</h2>
              <p>{productDesc(detail)}</p>
              <div className="shop-detail-price">{money(detail.priceCents, L.free)}</div>
              <div className="shop-card-rating">
                <Rate disabled allowHalf value={detail.rating} />
                <span>
                  {detail.rating.toFixed(1)} · {L.sold} {detail.soldCount.toLocaleString()}
                </span>
              </div>
              <div className="shop-detail-meta">
                {L.stock}: {detail.stock}
              </div>
              <Button
                type="primary"
                size="large"
                icon={<ShoppingCartOutlined />}
                disabled={detail.stock <= 0}
                loading={addingId === detail.id}
                onClick={() => void addToCart(detail)}
              >
                {detail.stock > 0 ? L.add : L.out}
              </Button>
            </div>
          </div>
        )}
        {detail && (
          <div className="shop-reviews">
            <h3>
              {L.reviews} <small>({detail.reviewCount})</small>
            </h3>
            {reviewsLoading ? (
              <Spin />
            ) : reviews.length === 0 ? (
              <Empty description={L.noReviews} image={Empty.PRESENTED_IMAGE_SIMPLE} />
            ) : (
              <div className="shop-review-list">
                {reviews.map((review) => (
                  <div key={review.id} className="shop-review-item">
                    <Avatar src={assetUrl(review.author?.avatarUrl)} size={34}>
                      {review.author?.nickname?.[0]}
                    </Avatar>
                    <div className="shop-review-body">
                      <div className="shop-review-head">
                        <b>{review.author?.nickname}</b>
                        <Rate disabled value={review.rating} />
                        <span>{new Date(review.createdAt).toLocaleDateString(i18n.language === 'zh' ? 'zh-CN' : 'en-US')}</span>
                      </div>
                      <p>{review.content}</p>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}
      </Modal>

      <Modal
        open={successOpen}
        title={L.successTitle}
        onCancel={() => setSuccessOpen(false)}
        footer={[
          <Button key="shop" onClick={() => setSuccessOpen(false)}>
            {L.continue}
          </Button>,
          <Button
            key="orders"
            type="primary"
            onClick={() => {
              setSuccessOpen(false);
              navigate('/orders');
            }}
          >
            {L.viewOrders}
          </Button>,
        ]}
      >
        <p>{L.successBody}</p>
        {lastOrder && (
          <Card size="small" className="shop-success-card">
            <div>
              <b>{lastOrder.orderNo}</b>
              <Tag color="gold" style={{ marginLeft: 8 }}>
                {statusLabel(lastOrder.status)}
              </Tag>
            </div>
            <div className="shop-success-total">{money(lastOrder.totalCents, L.free)}</div>
          </Card>
        )}
      </Modal>
    </div>
  );
}
