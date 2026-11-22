import { useEffect, useRef, useState } from 'react';
import {
  Alert,
  Button,
  Card,
  Col,
  Empty,
  Image,
  Input,
  List,
  Row,
  Select,
  Space,
  Spin,
  Tabs,
  Tag,
  Upload,
  message,
} from 'antd';
import {
  CameraOutlined,
  CopyOutlined,
  DownloadOutlined,
  PictureOutlined,
  SendOutlined,
  TranslationOutlined,
  UploadOutlined,
} from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { api, assetUrl, FALLBACK_IMAGE } from '../api/client';
import { useAuth } from '../context/AuthContext';

interface CopyResponse {
  content: string;
}

interface PosterCreateResponse {
  taskId: string;
}

interface PosterTask {
  taskId: string;
  status: 'pending' | 'running' | 'succeeded' | 'failed';
  imageUrl?: string;
  message?: string;
}

interface MenuOCRResponse {
  text: string;
}

interface TranslatedDish {
  name: string;
  translation: string;
  notes?: string;
}

interface MenuTranslateResponse {
  items: TranslatedDish[];
}

interface DishExplainResponse {
  title: string;
  content: string;
  tags?: string[];
}

interface ChatMessage {
  role: 'user' | 'assistant';
  content: string;
}

const COPY = {
  zh: {
    title: '商家 AI 工作台',
    subtitle: '面向本地文旅商家：文案、海报、菜单识别与客服，一座工作台完成内容生产。',
    copyTab: '营销文案',
    posterTab: 'AI 海报',
    menuTab: '拍照翻译',
    chatTab: '商家客服',
    topic: '中文主题 / 菜品 / 活动',
    topicPlaceholder: '例如：干蒸烧卖，皮薄馅足，广州早茶必点',
    type: '内容类型',
    tone: '文案风格',
    language: '输出语言',
    generateCopy: '生成文案',
    copyResult: '生成结果',
    copyEmpty: '填写主题后点击生成，AI 会返回一段可直接使用的营销文案。',
    copied: '已复制',
    posterPrompt: '海报提示词',
    posterPromptPlaceholder: '例如：广式早茶，虾饺与烧卖，暖色晨光，电影感美食摄影',
    posterStyle: '海报风格',
    generatePoster: '生成海报',
    posterHint: '海报由通义万相异步生成，通常需要 10–30 秒，请保持页面打开。',
    posterEmpty: '选择风格并生成，海报会显示在这里。',
    posterStatus: '任务状态',
    download: '下载海报',
    menuUpload: '上传菜单图片',
    menuHint: '支持 JPG / PNG，建议拍摄清晰的菜单页或菜品图。',
    recognize: '识别菜单',
    translate: '翻译菜单',
    ocrResult: '识别结果',
    ocrEmpty: '上传图片后点击识别，AI 会提取菜名与价格。',
    dishes: '菜品关键词',
    dishHint: '点击菜名生成文化卡片，或让 AI 服务员用简单语言解释。',
    explain: '详细解释',
    cardTitle: '文化卡片',
    translateResult: '菜单翻译',
    translateEmpty: '识别后选择目标语言，点击翻译。',
    chatWelcome: '你好，我是商家 AI 客服。可以问我营业时间、推荐菜、英文介绍或活动策划。',
    chatPlaceholder: '输入顾客常问的问题...',
    send: '发送',
    loginRequired: '登录后即可使用商家 AI 工作台。',
    failed: '请求失败，请稍后重试。',
    posterPending: '生成中',
    posterSucceeded: '已完成',
    posterFailed: '生成失败',
    posterIdle: '未开始',
    typeDish: '菜品',
    typeEvent: '活动',
    typeScenery: '景点',
    typeGeneral: '通用',
    toneFriendly: '亲切自然',
    toneLiterary: '文艺',
    tonePromotional: '促销转化',
    langZh: '中文',
    langEn: 'English',
    langJa: '日本語',
    langKo: '한국어',
    langEs: 'Español',
    langFr: 'Français',
    styleDefault: '默认',
    styleCartoon: '卡通',
    styleRealistic: '写实',
    styleCyberpunk: '赛博朋克',
    copyAction: '复制',
    quickQuestions: ['你们几点营业？', '招牌菜是什么？', '帮我写一段英文介绍', '设计一个周末促销活动'],
  },
  en: {
    title: 'Merchant AI Studio',
    subtitle: 'Marketing copy, posters, menu OCR and customer chat for local travel businesses.',
    copyTab: 'Marketing Copy',
    posterTab: 'AI Poster',
    menuTab: 'Photo Translate',
    chatTab: 'Merchant Chat',
    topic: 'Topic / dish / event',
    topicPlaceholder: 'e.g. Shrimp dumplings, thin skin, a Cantonese morning tea classic',
    type: 'Content type',
    tone: 'Tone',
    language: 'Output language',
    generateCopy: 'Generate Copy',
    copyResult: 'Result',
    copyEmpty: 'Enter a topic and generate ready-to-use marketing copy.',
    copied: 'Copied',
    posterPrompt: 'Poster prompt',
    posterPromptPlaceholder: 'e.g. Cantonese dim sum, warm morning light, cinematic food photography',
    posterStyle: 'Poster style',
    generatePoster: 'Generate Poster',
    posterHint: 'Wanx generates asynchronously, usually in 10–30 seconds. Keep this page open.',
    posterEmpty: 'Pick a style and generate; the poster will appear here.',
    posterStatus: 'Task status',
    download: 'Download poster',
    menuUpload: 'Upload menu image',
    menuHint: 'JPG / PNG. A clear menu page or dish photo works best.',
    recognize: 'Recognize Menu',
    translate: 'Translate Menu',
    ocrResult: 'OCR result',
    ocrEmpty: 'Upload an image and recognize to extract dish names and prices.',
    dishes: 'Dish keywords',
    dishHint: 'Click a dish for a culture card, or ask the AI waiter to explain it simply.',
    explain: 'Explain in detail',
    cardTitle: 'Culture card',
    translateResult: 'Menu translation',
    translateEmpty: 'Recognize first, pick a target language and translate.',
    chatWelcome: 'Hi, I am the merchant AI assistant. Ask me about opening hours, signature dishes, English intros or promotions.',
    chatPlaceholder: 'Type a common customer question...',
    send: 'Send',
    loginRequired: 'Sign in to use the Merchant AI Studio.',
    failed: 'Request failed, please retry.',
    posterPending: 'Generating',
    posterSucceeded: 'Done',
    posterFailed: 'Failed',
    posterIdle: 'Not started',
    typeDish: 'Dish',
    typeEvent: 'Event',
    typeScenery: 'Scenery',
    typeGeneral: 'General',
    toneFriendly: 'Friendly',
    toneLiterary: 'Literary',
    tonePromotional: 'Promotional',
    langZh: '中文',
    langEn: 'English',
    langJa: '日本語',
    langKo: '한국어',
    langEs: 'Español',
    langFr: 'Français',
    styleDefault: 'Default',
    styleCartoon: 'Cartoon',
    styleRealistic: 'Realistic',
    styleCyberpunk: 'Cyberpunk',
    copyAction: 'Copy',
    quickQuestions: ['What are your opening hours?', 'What is your signature dish?', 'Write an English introduction', 'Plan a weekend promotion'],
  },
} as const;

function readDataUrl(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => resolve(String(reader.result || ''));
    reader.onerror = () => reject(new Error('Could not read image'));
    reader.readAsDataURL(file);
  });
}

function extractDishes(text: string): string[] {
  const names = text
    .split(/\r?\n/)
    .map((line) => line.replace(/^[-*\d.\s]+/, '').split('|')[0].trim())
    .filter((line) => line.length > 0 && line.length < 40);
  return Array.from(new Set(names)).slice(0, 12);
}

export default function MerchantPage() {
  const { t, i18n } = useTranslation();
  const { user, openAuth } = useAuth();
  const L = i18n.language === 'zh' ? COPY.zh : COPY.en;

  const [copyTopic, setCopyTopic] = useState('');
  const [copyType, setCopyType] = useState('general');
  const [copyTone, setCopyTone] = useState('friendly');
  const [copyLang, setCopyLang] = useState(i18n.language === 'zh' ? 'zh' : 'en');
  const [copyBusy, setCopyBusy] = useState(false);
  const [copyResult, setCopyResult] = useState('');

  const [posterPrompt, setPosterPrompt] = useState('');
  const [posterStyle, setPosterStyle] = useState('realistic');
  const [posterBusy, setPosterBusy] = useState(false);
  const [posterTaskId, setPosterTaskId] = useState('');
  const [posterStatus, setPosterStatus] = useState<PosterTask['status'] | ''>('');
  const [posterUrl, setPosterUrl] = useState('');
  const [posterError, setPosterError] = useState('');

  const [menuImage, setMenuImage] = useState('');
  const [menuLanguage, setMenuLanguage] = useState('en');
  const [ocrText, setOcrText] = useState('');
  const [ocrBusy, setOcrBusy] = useState(false);
  const [translated, setTranslated] = useState<TranslatedDish[]>([]);
  const [translateBusy, setTranslateBusy] = useState(false);
  const [dish, setDish] = useState<DishExplainResponse | null>(null);
  const [dishExplain, setDishExplain] = useState('');
  const [dishBusy, setDishBusy] = useState('');
  const menuFileRef = useRef<File | null>(null);

  const [chat, setChat] = useState<ChatMessage[]>([{ role: 'assistant', content: L.chatWelcome }]);
  const [chatInput, setChatInput] = useState('');
  const [chatBusy, setChatBusy] = useState(false);
  const [sessionId, setSessionId] = useState<number | undefined>();
  const chatBottomRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    setChat((prev) =>
      prev.length === 1 && prev[0].role === 'assistant' ? [{ role: 'assistant', content: L.chatWelcome }] : prev,
    );
  }, [L.chatWelcome]);

  useEffect(() => {
    chatBottomRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [chat, chatBusy]);

  useEffect(() => {
    if (!posterTaskId) return;
    let cancelled = false;
    const poll = async () => {
      try {
        const task = await api.get<PosterTask>(`/api/ai/poster/${posterTaskId}`);
        if (cancelled) return;
        setPosterStatus(task.status);
        if (task.imageUrl) setPosterUrl(assetUrl(task.imageUrl));
        if (task.status === 'succeeded' || task.status === 'failed') {
          setPosterBusy(false);
          setPosterTaskId('');
          if (task.status === 'failed') {
            setPosterError(task.message || L.failed);
          } else {
            setPosterError('');
            message.success(L.posterSucceeded);
          }
        }
      } catch (e) {
        if (cancelled) return;
        setPosterBusy(false);
        setPosterTaskId('');
        setPosterError((e as Error).message || L.failed);
      }
    };
    void poll();
    const timer = setInterval(() => void poll(), 3000);
    return () => {
      cancelled = true;
      clearInterval(timer);
    };
  }, [posterTaskId, L.failed, L.posterSucceeded]);

  if (!user) {
    return (
      <Card className="merchant-login-card">
        <Empty description={L.loginRequired} />
        <Button type="primary" onClick={openAuth}>
          {t('common.login')}
        </Button>
      </Card>
    );
  }

  const generateCopy = async () => {
    if (!copyTopic.trim()) {
      message.warning(L.topic);
      return;
    }
    setCopyBusy(true);
    setCopyResult('');
    try {
      const res = await api.post<CopyResponse>('/api/ai/copy', {
        topic: copyTopic.trim(),
        type: copyType,
        tone: copyTone,
        language: copyLang,
      });
      setCopyResult(res.content);
    } catch (e) {
      message.error((e as Error).message || L.failed);
    } finally {
      setCopyBusy(false);
    }
  };

  const generatePoster = async () => {
    if (!posterPrompt.trim()) {
      message.warning(L.posterPrompt);
      return;
    }
    setPosterBusy(true);
    setPosterError('');
    setPosterUrl('');
    setPosterStatus('pending');
    try {
      const res = await api.post<PosterCreateResponse>('/api/ai/poster', {
        prompt: posterPrompt.trim(),
        style: posterStyle,
        language: copyLang,
      });
      setPosterTaskId(res.taskId);
    } catch (e) {
      setPosterBusy(false);
      setPosterStatus('failed');
      setPosterError((e as Error).message || L.failed);
    }
  };

  const onMenuFile = async (file: File) => {
    if (!file.type.startsWith('image/')) {
      message.error(t('common.onlyImages'));
      return;
    }
    if (file.size > 5 * 1024 * 1024) {
      message.error(t('common.imageTooLarge'));
      return;
    }
    menuFileRef.current = file;
    setMenuImage(await readDataUrl(file));
    setOcrText('');
    setTranslated([]);
    setDish(null);
    setDishExplain('');
  };

  const recognizeMenu = async () => {
    const file = menuFileRef.current;
    if (!file) {
      message.warning(L.menuUpload);
      return;
    }
    setOcrBusy(true);
    setDish(null);
    setDishExplain('');
    try {
      const image = await readDataUrl(file);
      const res = await api.post<MenuOCRResponse>('/api/ai/menu-ocr', {
        image,
        language: menuLanguage,
      });
      setOcrText(res.text);
    } catch (e) {
      message.error((e as Error).message || L.failed);
    } finally {
      setOcrBusy(false);
    }
  };

  const translateMenu = async () => {
    const file = menuFileRef.current;
    if (!file) {
      message.warning(L.menuUpload);
      return;
    }
    setTranslateBusy(true);
    try {
      const image = await readDataUrl(file);
      const res = await api.post<MenuTranslateResponse>('/api/ai/menu-translate', {
        image,
        language: menuLanguage,
      });
      setTranslated(res.items || []);
    } catch (e) {
      message.error((e as Error).message || L.failed);
    } finally {
      setTranslateBusy(false);
    }
  };

  const explainDish = async (name: string, kind: 'card' | 'explain') => {
    setDishBusy(name + kind);
    try {
      const res = await api.post<DishExplainResponse>('/api/ai/dish-explain', {
        name,
        kind,
        language: menuLanguage,
      });
      if (kind === 'card') {
        setDish(res);
        setDishExplain('');
      } else {
        setDishExplain(res.content);
      }
    } catch (e) {
      message.error((e as Error).message || L.failed);
    } finally {
      setDishBusy('');
    }
  };

  const sendChat = async (text?: string) => {
    const question = (text ?? chatInput).trim();
    if (!question || chatBusy) return;
    setChat((prev) => [...prev, { role: 'user', content: question }]);
    setChatInput('');
    setChatBusy(true);
    try {
      const res = await api.post<{ sessionId: number; reply: string }>('/api/guide/chat', {
        sessionId,
        question,
      });
      setSessionId(res.sessionId);
      setChat((prev) => [...prev, { role: 'assistant', content: res.reply }]);
    } catch (e) {
      setChat((prev) => [...prev, { role: 'assistant', content: (e as Error).message || L.failed }]);
    } finally {
      setChatBusy(false);
    }
  };

  const copyText = async () => {
    if (!copyResult) return;
    try {
      await navigator.clipboard.writeText(copyResult);
      message.success(L.copied);
    } catch {
      message.warning(L.copyAction);
    }
  };

  const dishes = extractDishes(ocrText);

  const copyTab = (
    <div className="merchant-panel">
      <Row gutter={[20, 20]}>
        <Col xs={24} lg={10}>
          <Card title={L.copyTab} className="merchant-form-card">
            <Space direction="vertical" size="middle" style={{ width: '100%' }}>
              <div>
                <div className="merchant-label">{L.topic}</div>
                <Input.TextArea
                  rows={5}
                  value={copyTopic}
                  onChange={(e) => setCopyTopic(e.target.value)}
                  placeholder={L.topicPlaceholder}
                  maxLength={500}
                  showCount
                />
              </div>
              <Row gutter={12}>
                <Col span={12}>
                  <div className="merchant-label">{L.type}</div>
                  <Select
                    value={copyType}
                    onChange={setCopyType}
                    style={{ width: '100%' }}
                    options={[
                      { value: 'general', label: L.typeGeneral },
                      { value: 'dish', label: L.typeDish },
                      { value: 'event', label: L.typeEvent },
                      { value: 'scenery', label: L.typeScenery },
                    ]}
                  />
                </Col>
                <Col span={12}>
                  <div className="merchant-label">{L.tone}</div>
                  <Select
                    value={copyTone}
                    onChange={setCopyTone}
                    style={{ width: '100%' }}
                    options={[
                      { value: 'friendly', label: L.toneFriendly },
                      { value: 'literary', label: L.toneLiterary },
                      { value: 'promotional', label: L.tonePromotional },
                    ]}
                  />
                </Col>
              </Row>
              <div>
                <div className="merchant-label">{L.language}</div>
                <Select
                  value={copyLang}
                  onChange={setCopyLang}
                  style={{ width: '100%' }}
                  options={[
                    { value: 'zh', label: L.langZh },
                    { value: 'en', label: L.langEn },
                  ]}
                />
              </div>
              <Button type="primary" block loading={copyBusy} onClick={() => void generateCopy()}>
                {L.generateCopy}
              </Button>
            </Space>
          </Card>
        </Col>
        <Col xs={24} lg={14}>
          <Card
            title={L.copyResult}
            className="merchant-result-card"
            extra={
              copyResult && (
                <Button size="small" icon={<CopyOutlined />} onClick={() => void copyText()}>
                  {L.copyAction}
                </Button>
              )
            }
          >
            {copyResult ? (
              <div className="merchant-copy">{copyResult}</div>
            ) : (
              <Empty description={L.copyEmpty} image={Empty.PRESENTED_IMAGE_SIMPLE} />
            )}
          </Card>
        </Col>
      </Row>
    </div>
  );

  const posterTab = (
    <div className="merchant-panel">
      <Row gutter={[20, 20]}>
        <Col xs={24} lg={10}>
          <Card title={L.posterTab} className="merchant-form-card">
            <Space direction="vertical" size="middle" style={{ width: '100%' }}>
              <div>
                <div className="merchant-label">{L.posterPrompt}</div>
                <Input.TextArea
                  rows={5}
                  value={posterPrompt}
                  onChange={(e) => setPosterPrompt(e.target.value)}
                  placeholder={L.posterPromptPlaceholder}
                  maxLength={500}
                  showCount
                />
              </div>
              <div>
                <div className="merchant-label">{L.posterStyle}</div>
                <Select
                  value={posterStyle}
                  onChange={setPosterStyle}
                  style={{ width: '100%' }}
                  options={[
                    { value: 'default', label: L.styleDefault },
                    { value: 'cartoon', label: L.styleCartoon },
                    { value: 'realistic', label: L.styleRealistic },
                    { value: 'cyberpunk', label: L.styleCyberpunk },
                  ]}
                />
              </div>
              <Button
                type="primary"
                block
                loading={posterBusy}
                icon={<PictureOutlined />}
                onClick={() => void generatePoster()}
              >
                {L.generatePoster}
              </Button>
              <Alert type="info" showIcon message={L.posterHint} />
            </Space>
          </Card>
        </Col>
        <Col xs={24} lg={14}>
          <Card
            title={L.posterTab}
            className="merchant-result-card merchant-poster-card"
            extra={
              posterUrl && (
                <Button
                  size="small"
                  icon={<DownloadOutlined />}
                  href={posterUrl}
                  target="_blank"
                  rel="noreferrer"
                >
                  {L.download}
                </Button>
              )
            }
          >
            {posterError && <Alert type="error" showIcon message={posterError} style={{ marginBottom: 12 }} />}
            {posterBusy && (
              <div className="merchant-poster-loading">
                <Spin />
                <span>
                  {L.posterStatus}: {posterStatus === 'running' ? L.posterPending : L.posterPending}
                </span>
              </div>
            )}
            {posterUrl ? (
              <Image src={posterUrl} fallback={FALLBACK_IMAGE} alt="AI poster" className="merchant-poster" />
            ) : (
              !posterBusy && <Empty description={L.posterEmpty} image={Empty.PRESENTED_IMAGE_SIMPLE} />
            )}
          </Card>
        </Col>
      </Row>
    </div>
  );

  const menuTab = (
    <div className="merchant-panel">
      <Row gutter={[20, 20]}>
        <Col xs={24} lg={9}>
          <Card title={L.menuTab} className="merchant-form-card">
            <Space direction="vertical" size="middle" style={{ width: '100%' }}>
              <Upload
                accept="image/*"
                showUploadList={false}
                beforeUpload={(file) => {
                  void onMenuFile(file as File);
                  return false;
                }}
              >
                <div className={`merchant-upload-box${menuImage ? ' has-image' : ''}`}>
                  {menuImage ? (
                    <img src={menuImage} alt="menu preview" />
                  ) : (
                    <>
                      <CameraOutlined />
                      <span>{L.menuUpload}</span>
                      <small>{L.menuHint}</small>
                    </>
                  )}
                </div>
              </Upload>
              <div>
                <div className="merchant-label">{L.language}</div>
                <Select
                  value={menuLanguage}
                  onChange={setMenuLanguage}
                  style={{ width: '100%' }}
                  options={[
                    { value: 'en', label: L.langEn },
                    { value: 'ja', label: L.langJa },
                    { value: 'ko', label: L.langKo },
                    { value: 'es', label: L.langEs },
                    { value: 'fr', label: L.langFr },
                    { value: 'zh', label: L.langZh },
                  ]}
                />
              </div>
              <Space wrap>
                <Button type="primary" loading={ocrBusy} onClick={() => void recognizeMenu()}>
                  {L.recognize}
                </Button>
                <Button
                  icon={<TranslationOutlined />}
                  loading={translateBusy}
                  onClick={() => void translateMenu()}
                >
                  {L.translate}
                </Button>
              </Space>
            </Space>
          </Card>
        </Col>
        <Col xs={24} lg={15}>
          <Space direction="vertical" size="middle" style={{ width: '100%' }}>
            <Card title={L.ocrResult} className="merchant-result-card">
              {ocrText ? (
                <>
                  <pre className="merchant-ocr">{ocrText}</pre>
                  {dishes.length > 0 && (
                    <>
                      <div className="merchant-label" style={{ marginTop: 14 }}>
                        {L.dishes}
                      </div>
                      <div className="merchant-chips">
                        {dishes.map((name) => (
                          <Tag
                            key={name}
                            color="blue"
                            className="merchant-chip"
                            onClick={() => void explainDish(name, 'card')}
                          >
                            {name}
                          </Tag>
                        ))}
                      </div>
                      <small className="merchant-hint">{L.dishHint}</small>
                    </>
                  )}
                </>
              ) : (
                <Empty description={L.ocrEmpty} image={Empty.PRESENTED_IMAGE_SIMPLE} />
              )}
            </Card>

            {(dish || dishExplain) && (
              <Card
                title={L.cardTitle}
                className="merchant-result-card"
                extra={
                  dish && (
                    <Button
                      size="small"
                      loading={dishBusy === dish.title + 'explain'}
                      onClick={() => void explainDish(dish.title, 'explain')}
                    >
                      {L.explain}
                    </Button>
                  )
                }
              >
                {dish && <h3 className="merchant-dish-title">{dish.title}</h3>}
                {dish && <p>{dish.content}</p>}
                {dish?.tags?.map((tag) => (
                  <Tag key={tag}>{tag}</Tag>
                ))}
                {dishBusy.endsWith('card') && <Spin size="small" />}
                {dishExplain && <div className="merchant-copy">{dishExplain}</div>}
              </Card>
            )}

            {translated.length > 0 && (
              <Card title={L.translateResult} className="merchant-result-card">
                <List
                  dataSource={translated}
                  renderItem={(item) => (
                    <List.Item>
                      <List.Item.Meta
                        title={
                          <span>
                            {item.name} <span className="merchant-arrow">→</span> {item.translation}
                          </span>
                        }
                        description={item.notes}
                      />
                    </List.Item>
                  )}
                />
              </Card>
            )}
          </Space>
        </Col>
      </Row>
    </div>
  );

  const chatTab = (
    <div className="merchant-panel">
      <Card
        title={L.chatTab}
        className="merchant-chat-card"
        extra={<Tag color="green">AI</Tag>}
      >
        <div className="merchant-chat-box">
          {chat.map((m, i) => (
            <div key={i} className={`merchant-chat-row ${m.role}`}>
              <div className="merchant-chat-bubble">{m.content}</div>
            </div>
          ))}
          {chatBusy && (
            <div className="merchant-chat-row assistant">
              <div className="merchant-chat-bubble">
                <Spin size="small" />
              </div>
            </div>
          )}
          <div ref={chatBottomRef} />
        </div>
        <div className="merchant-chat-quick">
          {L.quickQuestions.map((q) => (
            <Button key={q} size="small" onClick={() => void sendChat(q)} disabled={chatBusy}>
              {q}
            </Button>
          ))}
        </div>
        <Space.Compact style={{ width: '100%', marginTop: 12 }}>
          <Input
            value={chatInput}
            onChange={(e) => setChatInput(e.target.value)}
            placeholder={L.chatPlaceholder}
            onPressEnter={() => void sendChat()}
            disabled={chatBusy}
          />
          <Button type="primary" icon={<SendOutlined />} loading={chatBusy} onClick={() => void sendChat()}>
            {L.send}
          </Button>
        </Space.Compact>
      </Card>
    </div>
  );

  return (
    <div className="merchant-page">
      <div className="merchant-hero">
        <div>
          <h1>{L.title}</h1>
          <p>{L.subtitle}</p>
        </div>
        <Tag color="blue" icon={<UploadOutlined />}>
          {i18n.language === 'zh' ? '阿里云百炼 · qwen-plus' : 'Aliyun Bailian · qwen-plus'}
        </Tag>
      </div>
      <Tabs
        defaultActiveKey="copy"
        items={[
          { key: 'copy', label: L.copyTab, children: copyTab },
          { key: 'poster', label: L.posterTab, children: posterTab },
          { key: 'menu', label: L.menuTab, children: menuTab },
          { key: 'chat', label: L.chatTab, children: chatTab },
        ]}
      />
    </div>
  );
}
