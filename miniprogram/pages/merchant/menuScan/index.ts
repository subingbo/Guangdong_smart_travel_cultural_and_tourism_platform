import { getLanguage, getTranslations } from '../../../utils/i18n';
import { dishExplain, menuOCR, menuTranslate } from '../../../services/api';
import { getToken } from '../../../utils/auth';
import { gotoLogin } from '../../../utils/request';
import { readImageBase64 } from '../../../utils/upload';

Page({
  data: {
    t: {} as any,
    imageUrl: '',
    translatedImageUrl: '',
    loading: false,
    translating: false,
    error: '',
    overlayText: '',
    rawOCRItems: [] as any[],
    ocrItems: [] as any[],
    showExplanation: false,
    explanation: {
      title: '',
      content: '',
      tags: [] as string[]
    },
    imageInfo: { width: 0, height: 0 },
    containerInfo: { width: 0, height: 0 },
    
    // New Data for Culture Card
    showWordCloud: false,
    keywords: [] as any[],
    translations: [] as any[],
    cultureCard: null as any,
    showCultureCard: false,
    cardLoading: false,
    cardStreaming: false,
    cardStreamText: '',
    imageDataB64: '', // Cache of the base64 payload sent to AI endpoints
    analyzing: false,
    analysisStep: '',
    
    // Card Resize Data
    cardHeight: 0,
    cardStartHeight: 0,
    cardStartY: 0,

    // Drawer Data
    drawerHeight: 0, // Current height in px
    startHeight: 0,
    startY: 0,
    minHeight: 150, // Minimum visible height (handle + some content)
    maxHeight: 0,   // Set in onLoad based on screen
    isDrawerOpen: false
  },
  
  onLoad() {
    const sysInfo = wx.getSystemInfoSync();
    const maxHeight = sysInfo.windowHeight * 0.8; // Max 80% of screen
    this.setData({ 
        maxHeight,
        drawerHeight: 0 // Start hidden
    });
  },

  onShow() {
    const lang = getLanguage();
    this.setData({ t: getTranslations(lang) });
    wx.setNavigationBarTitle({
      title: lang === 'en' ? 'Menu Translator' : '菜单翻译'
    });
  },

  onBack() {
    wx.navigateBack();
  },

  async onChooseAlbum() {
    this.chooseImage(['album']);
  },

  async onTakePhoto() {
    const ctx = wx.createCameraContext();
    ctx.takePhoto({
      quality: 'high',
      success: async (res) => {
        const tempPath = res.tempImagePath;
        try {
          const imgInfo = await wx.getImageInfo({ src: tempPath });
          this.setData({
            imageInfo: { width: imgInfo.width, height: imgInfo.height }
          });
        } catch (e) {
          console.error('Get image info failed', e);
        }

        this.setData({
          imageUrl: tempPath,
          translatedImageUrl: '',
          error: '',
          overlayText: '',
          ocrItems: [],
          rawOCRItems: [],
          showExplanation: false,
          showWordCloud: false,
          showCultureCard: false,
          cardLoading: false,
          cardStreaming: false,
          cardStreamText: '',
          isDrawerOpen: false,
          drawerHeight: 0,
          imageDataB64: '',
          translations: []
        });
      },
      fail: (e) => {
        console.error('Take photo failed', e);
        wx.showToast({ title: 'Capture failed', icon: 'none' });
      }
    });
  },

  // Helper to ensure the current image is available as a base64 payload.
  async ensureImageBase64(): Promise<string> {
    if (this.data.imageDataB64) return this.data.imageDataB64;
    if (!this.data.imageUrl) throw new Error('No image selected');

    wx.showLoading({ title: 'Preparing...' });
    try {
      const b64 = await readImageBase64(this.data.imageUrl);
      this.setData({ imageDataB64: b64 });
      return b64;
    } finally {
      wx.hideLoading();
    }
  },

  requireLogin(): boolean {
    if (getToken()) return true;
    wx.showModal({
      title: '需要登录',
      content: '菜单识别与翻译需要先登录，是否前往登录页？',
      confirmText: '去登录',
      success: (res) => {
        if (res.confirm) gotoLogin();
      }
    });
    return false;
  },

  async onCultureCard() {
    if (!this.data.imageUrl) {
      wx.showToast({ title: 'Please select an image first', icon: 'none' });
      return;
    }
    if (!this.requireLogin()) return;

    this.setData({
        isDrawerOpen: true,
        drawerHeight: this.data.maxHeight * 0.5,
        analyzing: true,
        analysisStep: '正在识别文字...',
        keywords: [],
        translations: [] // Clear previous results
    });

    try {
      const image = await this.ensureImageBase64();

      // OCR the image through the local backend (qwen-vl).
      const ocrResult = await menuOCR(image, getLanguage());
      const text = ocrResult.text || '';

      // Derive tappable dish keywords from the OCR text client-side.
      this.setData({ analysisStep: '正在提取菜品关键词...' });
      const keywords = this.buildKeywordsFromOcr(text);

      this.setData({
        keywords,
        analyzing: false,
        analysisStep: ''
      });

    } catch (e: any) {
      console.error('Culture analysis failed', e);
      wx.showToast({ title: 'Analysis Failed', icon: 'none' });
      this.setData({
          analyzing: false,
          analysisStep: '分析失败',
          isDrawerOpen: false
      });
    }
  },

  buildKeywordsFromOcr(text: string) {
    const lines = text.split(/\n+/).map((s) => s.trim()).filter(Boolean);
    const seen = new Set<string>();
    const keywords: any[] = [];
    for (const line of lines) {
      // menuOCR lines look like "name | price"; keep the name part only.
      const word = line.split('|')[0].trim();
      if (!word || word.length > 20 || seen.has(word)) continue;
      seen.add(word);
      keywords.push({
        word,
        category: 'Food',
        color: this.getCategoryColor('Food')
      });
    }
    return keywords.slice(0, 12);
  },

  getCategoryColor(category: string) {
    const colors: any = {
      'Food': '#FF9C6E',
      'Architecture': '#5CDBD3',
      'Custom': '#FFC069',
      'Other': '#d9d9d9'
    };
    return colors[category] || colors['Other'];
  },

  async onKeywordTap(e: any) {
    const keyword = e.currentTarget.dataset.keyword;

    try {
      this.setData({
        showCultureCard: true,
        cardLoading: true,
        cardStreaming: false,
        cardStreamText: '',
        cultureCard: {
          title: keyword,
          description: '',
          cultural_significance: '',
          tags: []
        },
        cardHeight: 0 // Set to 0 for auto height
      });

      // The backend generates the whole card synchronously (qwen text model).
      const card = await dishExplain(keyword, 'card', getLanguage());

      this.setData({
        cardLoading: false,
        cultureCard: {
          title: card.title || keyword,
          description: card.content || '',
          cultural_significance: '',
          tags: card.tags || []
        }
      });
    } catch (e) {
      console.error('Card generation failed', e);
      wx.showToast({ title: 'Generation Failed', icon: 'none' });
      this.setData({ cardStreaming: false, cardLoading: false });
    }
  },

  onCloseCultureCard() {
    this.setData({ showCultureCard: false, cardStreaming: false, cardStreamText: '' });
  },

  onCloseWordCloud() {
    this.setData({ isDrawerOpen: false, drawerHeight: 0 });
  },

  async playCultureCardStream(chunks: string[]) {
    if (!chunks.length) return;
    await new Promise<void>((resolve) => {
      let index = 0;
      let text = '';
      const step = () => {
        if (index >= chunks.length) {
          this.setData({ cardStreamText: text });
          resolve();
          return;
        }
        const chunk = chunks[index];
        index += 1;
        if (chunk) {
          text += chunk;
          this.setData({ cardStreamText: text });
        }
        setTimeout(step, 20);
      };
      step();
    });
  },

  // Card Resize Logic
  onCardDragStart(e: any) {
    const clientY = e.touches[0].clientY;
    if (!this.data.cardHeight) {
      // If cardHeight is 0 (auto), get the current height first
      const query = wx.createSelectorQuery();
      query.select('.culture-card').boundingClientRect((rect) => {
        if (rect) {
          this.setData({
            cardHeight: rect.height,
            cardStartHeight: rect.height,
            cardStartY: clientY
          });
        }
      }).exec();
    } else {
      this.setData({
        cardStartY: clientY,
        cardStartHeight: this.data.cardHeight
      });
    }
  },

  onCardDragMove(e: any) {
    if (!this.data.cardStartHeight) return;
    
    const deltaY = e.touches[0].clientY - this.data.cardStartY; // Dragging DOWN increases height (if pulling bottom)
    // Since we are dragging the bottom edge, positive delta means increasing height
    let newHeight = this.data.cardStartHeight + deltaY;

    // Constraints
    const minHeight = 200;
    const maxHeight = wx.getSystemInfoSync().windowHeight * 0.9;

    if (newHeight < minHeight) newHeight = minHeight;
    if (newHeight > maxHeight) newHeight = maxHeight;

    this.setData({
      cardHeight: newHeight
    });
  },

  // Drawer Drag Logic
  onDragStart(e: any) {
    this.setData({
      startY: e.touches[0].clientY,
      startHeight: this.data.drawerHeight
    });
  },

  onDragMove(e: any) {
    const deltaY = this.data.startY - e.touches[0].clientY; // Dragging UP is positive delta
    let newHeight = this.data.startHeight + deltaY;

    // Clamp
    if (newHeight < this.data.minHeight) newHeight = this.data.minHeight;
    if (newHeight > this.data.maxHeight) newHeight = this.data.maxHeight;

    this.setData({
      drawerHeight: newHeight
    });
  },

  onDragEnd() {
    // Snap to the minimum height when the drawer is dragged close to it.
    if (this.data.drawerHeight < this.data.minHeight + 50) {
       this.setData({ drawerHeight: this.data.minHeight });
    }
  },

  async onTapOCRBox(e: any) {
    const text = e.currentTarget.dataset.text;
    if (!text) return;

    this.setData({
      showExplanation: true,
      explanation: { title: 'Thinking...', content: 'Asking AI for cultural details...', tags: [] }
    });

    try {
      const explain = await dishExplain(text, 'explain', getLanguage());
      this.setData({
        explanation: {
          title: explain.title || text,
          content: explain.content || '',
          tags: (explain.tags || []) as string[]
        }
      });
    } catch (e) {
      console.error('Explanation failed', e);
      this.setData({
        explanation: { title: 'Error', content: 'Failed to load explanation.', tags: [] }
      });
    }
  },

  onCloseExplanation() {
    this.setData({ showExplanation: false });
  },

  onResetCamera() {
    this.setData({
      imageUrl: '',
      translatedImageUrl: '',
      error: '',
      overlayText: '',
      showWordCloud: false,
      showCultureCard: false,
      cardLoading: false,
      cardStreaming: false,
      cardStreamText: '',
      isDrawerOpen: false,
      drawerHeight: 0,
      imageDataB64: '',
      translations: []
    });
  },

  onCameraError(e: any) {
    console.error('Camera error:', e);
    wx.showToast({ title: 'Camera access denied', icon: 'none' });
  },

  async onSaveImage() {
    if (!this.data.imageUrl) return;

    const itemList: string[] = [];
    
    if (this.data.translatedImageUrl) {
      itemList.push('Save Original');
      itemList.push('Save Translated');
      itemList.push('Save Both');
    } else {
      itemList.push('Save Original');
    }

    try {
      const res = await wx.showActionSheet({ itemList });
      const selected = itemList[res.tapIndex];
      const tasks: Promise<any>[] = [];

      if (selected === 'Save Original' || selected === 'Save Both') {
        if (this.data.imageUrl) tasks.push(this.saveToAlbum(this.data.imageUrl));
      }
      
      if (selected === 'Save Translated' || selected === 'Save Both') {
        if (this.data.translatedImageUrl) tasks.push(this.saveToAlbum(this.data.translatedImageUrl));
      }

      await Promise.all(tasks);
      wx.showToast({ title: 'Saved', icon: 'success' });
      
    } catch (e) {
      console.error('Save cancelled or failed', e);
    }
  },

  async saveToAlbum(filePath: string) {
    if (filePath.startsWith('http') && !filePath.startsWith('wxfile')) {
      try {
        const tempFilePath = await new Promise<string>((resolve, reject) => {
          wx.downloadFile({
            url: filePath,
            success: (res) => {
              if (res.statusCode === 200) resolve(res.tempFilePath);
              else reject(new Error(`Download failed with status ${res.statusCode}`));
            },
            fail: (err) => reject(err)
          });
        });
        filePath = tempFilePath;
      } catch (e) {
        console.error('Download failed', e);
        throw new Error('Download failed');
      }
    }

    try {
      await wx.saveImageToPhotosAlbum({ filePath });
    } catch (e: any) {
      if (e.errMsg && (e.errMsg.includes('auth deny') || e.errMsg.includes('authorize:fail'))) {
        const modalRes = await wx.showModal({
          title: 'Permission Required',
          content: 'Please allow access to save photos',
          confirmText: 'Settings'
        });
        if (modalRes.confirm) wx.openSetting();
      }
      throw e;
    }
  },

  async chooseImage(sourceType: Array<'album' | 'camera'>) {
    try {
      const res = await wx.chooseMedia({
        count: 1,
        mediaType: ['image'],
        sourceType: sourceType,
        sizeType: ['compressed'],
      });
      
      const tempFilePath = res.tempFiles[0].tempFilePath;
      
      try {
        const imgInfo = await wx.getImageInfo({ src: tempFilePath });
        this.setData({
          imageInfo: { width: imgInfo.width, height: imgInfo.height }
        });
      } catch (e) {
        console.error('Get image info failed', e);
      }

      this.setData({ 
        imageUrl: tempFilePath,
        translatedImageUrl: '',
        error: '',
        overlayText: '',
        ocrItems: [],
        rawOCRItems: [],
        showExplanation: false,
        showWordCloud: false,
        showCultureCard: false,
        cardLoading: false,
        cardStreaming: false,
        cardStreamText: '',
        isDrawerOpen: false,
        drawerHeight: 0,
        imageDataB64: '',
        translations: []
      });

    } catch (e) {
      console.error('Choose image cancelled or failed', e);
    }
  },

  async onTranslate() {
    if (!this.data.imageUrl) {
      wx.showToast({ title: 'Please select an image first', icon: 'none' });
      return;
    }
    if (this.data.translating) return;
    if (!this.requireLogin()) return;

    const langs = ['English', '中文', 'Japanese', 'Korean', 'Spanish', 'French'];
    const langCodes = ['en', 'zh', 'ja', 'ko', 'es', 'fr'];

    try {
      const res = await wx.showActionSheet({ itemList: langs });
      const targetLang = langCodes[res.tapIndex];
      await this.startTranslation(targetLang);
    } catch {
      // The user dismissed the action sheet.
    }
  },

  async startTranslation(targetLang: string) {
    this.setData({ translating: true, error: '' });
    wx.showLoading({ title: 'Translating...' });

    try {
      const image = await this.ensureImageBase64();
      // The backend returns the translated dish list synchronously.
      const result = await menuTranslate(image, targetLang);
      const translations = (result.items || []).map((item: any) => ({
        name: item.name,
        translation: item.translation,
        notes: item.notes || ''
      }));

      this.setData({
        translations,
        keywords: [],
        isDrawerOpen: true,
        drawerHeight: this.data.maxHeight * 0.5,
        analyzing: false,
        analysisStep: ''
      });

      if (translations.length === 0) {
        wx.showToast({ title: 'No dishes found', icon: 'none' });
      }
    } catch (err: any) {
      console.error(err);
      this.setData({ error: err.message || 'Error occurred' });
      wx.showToast({ title: 'Translation Failed', icon: 'none' });
    } finally {
      this.setData({ translating: false });
      wx.hideLoading();
    }
  }
});
