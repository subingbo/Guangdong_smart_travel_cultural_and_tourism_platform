import { useCallback, useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { Button, Card, Empty, Spin, Tag, message } from 'antd';
import { ArrowLeftOutlined, ShoppingOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { api, Order, Page } from '../api/client';
import { useAuth } from '../context/AuthContext';
import SafeImage from '../components/SafeImage';

const COPY = {
  zh: {
    title: '我的订单',
    subtitle: '预订单可在到店时出示订单号结算。',
    empty: '还没有订单，去商城逛逛吧。',
    shop: '去商城',
    orderNo: '订单号',
    total: '合计',
    pending: '待结算',
    paid: '已支付',
    completed: '已完成',
    canceled: '已取消',
    items: '件商品',
    free: '免费',
  },
  en: {
    title: 'My Orders',
    subtitle: 'Show your order number at the venue to settle.',
    empty: 'No orders yet. Visit the market.',
    shop: 'Go to shop',
    orderNo: 'Order',
    total: 'Total',
    pending: 'Pending',
    paid: 'Paid',
    completed: 'Completed',
    canceled: 'Canceled',
    items: 'items',
    free: 'Free',
  },
} as const;

function money(cents: number, free: string) {
  return cents <= 0 ? free : `¥${(cents / 100).toFixed(cents % 100 === 0 ? 0 : 2)}`;
}

export default function OrdersPage() {
  const { i18n } = useTranslation();
  const { user, openAuth } = useAuth();
  const L = i18n.language === 'zh' ? COPY.zh : COPY.en;
  const [orders, setOrders] = useState<Order[]>([]);
  const [loading, setLoading] = useState(true);

  const load = useCallback(async () => {
    if (!user) {
      setOrders([]);
      setLoading(false);
      return;
    }
    setLoading(true);
    try {
      const res = await api.get<Page<Order>>('/api/shop/orders?page=1&pageSize=20');
      setOrders(res.items || []);
    } catch (e) {
      message.error((e as Error).message);
      setOrders([]);
    } finally {
      setLoading(false);
    }
  }, [user?.id]);

  useEffect(() => {
    void load();
  }, [load]);

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

  const statusColor = (status: Order['status']) => {
    switch (status) {
      case 'paid':
        return 'blue';
      case 'completed':
        return 'green';
      case 'canceled':
        return 'default';
      default:
        return 'gold';
    }
  };

  if (!user) {
    return (
      <Card className="orders-login-card">
        <Empty description={i18n.language === 'zh' ? '登录后查看订单。' : 'Sign in to view your orders.'} />
        <Button type="primary" onClick={openAuth}>
          {i18n.language === 'zh' ? '登录' : 'Sign in'}
        </Button>
      </Card>
    );
  }

  return (
    <div className="orders-page">
      <div className="orders-hero">
        <div>
          <Link to="/shop" className="orders-back">
            <ArrowLeftOutlined /> {L.shop}
          </Link>
          <h1>{L.title}</h1>
          <p>{L.subtitle}</p>
        </div>
        <Button type="primary" icon={<ShoppingOutlined />} onClick={() => void load()}>
          {i18n.language === 'zh' ? '刷新' : 'Refresh'}
        </Button>
      </div>

      {loading ? (
        <div className="orders-loading">
          <Spin />
        </div>
      ) : orders.length === 0 ? (
        <Empty description={L.empty}>
          <Link to="/shop">
            <Button type="primary">{L.shop}</Button>
          </Link>
        </Empty>
      ) : (
        <div className="orders-list">
          {orders.map((order) => (
            <Card key={order.id} className="orders-card">
              <div className="orders-head">
                <div>
                  <small>{L.orderNo}</small>
                  <b>{order.orderNo}</b>
                  <div className="orders-date">
                    {new Date(order.createdAt).toLocaleString(i18n.language === 'zh' ? 'zh-CN' : 'en-US')}
                  </div>
                </div>
                <Tag color={statusColor(order.status)}>{statusLabel(order.status)}</Tag>
              </div>
              <div className="orders-items">
                {order.items.map((item) => (
                  <div key={item.productId} className="orders-item">
                    <SafeImage
                      className="orders-thumb"
                      src={item.imageUrl}
                      alt={i18n.language === 'zh' ? item.nameCn : item.nameEn}
                    />
                    <div className="orders-item-info">
                      <b>{i18n.language === 'zh' ? item.nameCn : item.nameEn}</b>
                      <span>
                        {money(item.priceCents, L.free)} × {item.quantity}
                      </span>
                    </div>
                  </div>
                ))}
              </div>
              <div className="orders-total">
                <span>
                  {order.items.reduce((n, item) => n + item.quantity, 0)} {L.items}
                </span>
                <b>
                  {L.total} {money(order.totalCents, L.free)}
                </b>
              </div>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}
