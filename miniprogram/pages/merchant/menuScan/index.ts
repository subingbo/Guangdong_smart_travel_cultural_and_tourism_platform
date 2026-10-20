import { getLanguage, getTranslations } from '../../../utils/i18n';

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
    cultureCard: null as any,
    showCultureCard: false,
    cardLoading: false,
    cardStreaming: false,
    cardStreamText: '',
    uploadedFileID: '', // Cache fileID
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
          uploadedFileID: ''
        });
      },
      fail: (e) => {
        console.error('Take photo failed', e);
        wx.showToast({ title: 'Capture failed', icon: 'none' });
      }
    });
  },

  // Helper to ensure image is uploaded
  async ensureImageUploaded(): Promise<string> {
    if (this.data.uploadedFileID) return this.data.uploadedFileID;
    if (!this.data.imageUrl) throw new Error('No image selected');

    wx.showLoading({ title: 'Uploading...' });
    const filePath = this.data.imageUrl;
    const cloudPath = `menu_images/${Date.now()}-${Math.random().toString(36).slice(-6)}.jpg`;
    
    try {
      const uploadRes = await wx.cloud.uploadFile({
        cloudPath,
        filePath
      });
      this.setData({ uploadedFileID: uploadRes.fileID });
      return uploadRes.fileID;
    } catch (e) {
      console.error('Upload failed', e);
      throw e;
    } finally {
      wx.hideLoading();
    }
  },

  async runOCR() {
    // ... existing runOCR logic if needed ...
  },

  // New: Culture Card Flow
  async onCultureCard() {
    if (!this.data.imageUrl) {
      wx.showToast({ title: 'Please select an image first', icon: 'none' });
      return;
    }

    // wx.showLoading({ title: 'Analyzing Culture...' });
    this.setData({
        isDrawerOpen: true,
        drawerHeight: this.data.maxHeight * 0.5,
        analyzing: true,
        analysisStep: '正在识别文字...',
        keywords: [] // Clear previous results
    });

    try {
      const fileID = await this.ensureImageUploaded();

      // 1. Run Qwen-OCR
      const ocrRes = await wx.cloud.callFunction({
        name: 'aiService',
        data: {
          action: 'qwen_ocr',
          payload: { fileID }
        }
      });
      const ocrResult = ocrRes.result as any;
      if (!ocrResult.success) throw new Error(ocrResult.message || 'OCR Failed');
      const text = ocrResult.text;

      // 2. Extract Keywords
      this.setData({ analysisStep: '正在提取文化关键词...' });
      const keywordRes = await wx.cloud.callFunction({
        name: 'aiService',
        data: {
          action: 'extract_keywords',
          payload: { text }
        }
      });
      const keywordResult = keywordRes.result as any;
      if (!keywordResult.success) throw new Error(keywordResult.message || 'Analysis Failed');

      // Process keywords for display
      let keywords = keywordResult.data.keywords.map((k: any) => ({
        ...k,
        color: this.getCategoryColor(k.category)
      }));

      // Sort: Cultural categories first
      const culturalCategories = ['Food', 'Architecture', 'Custom'];
      keywords.sort((a: any, b: any) => {
        const aIsCultural = culturalCategories.includes(a.category);
        const bIsCultural = culturalCategories.includes(b.category);
        if (aIsCultural && !bIsCultural) return -1;
        if (!aIsCultural && bIsCultural) return 1;
        return 0;
      });

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
          isDrawerOpen: false // Close drawer on error? Or keep open to show error? Closing for now.
      });
    } finally {
      // wx.hideLoading();
    }
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

      const startRes = await wx.cloud.callFunction({
        name: 'aiService',
        data: {
          action: 'generate_culture_card',
          payload: { keyword }
        }
      });
      const startResult = startRes.result as any;
      if (!startResult.success) throw new Error(startResult.message || 'Start Failed');
      
      // Stop loading spinner
      this.setData({ cardLoading: false, cardStreaming: false });

      const data = startResult.data;
      if (data) {
        this.setData({
          cultureCard: data,
          cardStreaming: false
        });
      } else {
        this.setData({
          cardStreaming: false
        });
      }
    } catch (e) {
      console.error('Card generation failed', e);
      wx.showToast({ title: 'Generation Failed', icon: 'none' });
      this.setData({ cardStreaming: false, cardLoading: false });
    } finally {
    }
  },

  async pollCultureCardResult(taskId: string) {
    const maxRetries = 30;
    let retries = 0;
    let interval = 1000;
    while (retries < maxRetries) {
      const pollRes = await wx.cloud.callFunction({
        name: 'aiService',
        data: {
          action: 'get_culture_card_result',
          payload: { taskId }
        }
      });
      const pollResult = pollRes.result as any;
      if (pollResult && pollResult.status === 'SUCCEEDED' && pollResult.data) {
        return pollResult.data;
      }
      if (pollResult && pollResult.status === 'FAILED') {
        throw new Error(pollResult.message || 'Generation Failed');
      }
      await new Promise((resolve) => setTimeout(resolve, interval));
      retries += 1;
      if (retries === 10) interval = 2000;
      if (retries === 20) interval = 3000;
    }
    throw new Error('Generation Timed Out');
  },

  onCloseCultureCard() {
    this.setData({ showCultureCard: false, cardStreaming: false, cardStreamText: '' });
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
    // Optional: Snap to nearest anchor (min, mid, max)
    // For now, leave as is or snap to min if too small
    if (this.data.drawerHeight < this.data.minHeight + 50) {
       // Close or snap to min? Let's snap to min
       this.setData({ drawerHeight: this.data.minHeight });
    }
  },

  // ... (Existing OCR & Translate functions: updateOCRLayout, calculateOCRStyles, onTapOCRBox, onCloseExplanation, onResetCamera, onCameraError, onSaveImage, saveToAlbum, chooseImage, onTranslate, startTranslation)
  
  updateOCRLayout() {
    const query = wx.createSelectorQuery();
    query.select('.camera-container').boundingClientRect((rect) => {
      if (rect && rect.width && rect.height) {
        this.setData({
          containerInfo: { width: rect.width, height: rect.height }
        });
        this.calculateOCRStyles();
      }
    }).exec();
  },

  calculateOCRStyles() {
    const { rawOCRItems, imageInfo, containerInfo } = this.data;
    if (!rawOCRItems.length || !imageInfo.width || !containerInfo.width) return;

    const imgW = imageInfo.width;
    const imgH = imageInfo.height;
    const conW = containerInfo.width;
    const conH = containerInfo.height;

    // AspectFit logic
    const scale = Math.min(conW / imgW, conH / imgH);
    const displayW = imgW * scale;
    const displayH = imgH * scale;
    const offsetX = (conW - displayW) / 2;
    const offsetY = (conH - displayH) / 2;

    const ocrItems = rawOCRItems.map((item: any) => {
      const box = item.box;
      const x = box.left_top.x;
      const y = box.left_top.y;
      const w = box.right_top.x - box.left_top.x;
      const h = box.left_bottom.y - box.left_top.y;

      const style = `
        left: ${x * scale + offsetX}px;
        top: ${y * scale + offsetY}px;
        width: ${w * scale}px;
        height: ${h * scale}px;
      `;

      return { ...item, style };
    });

    this.setData({ ocrItems });
  },

  async onTapOCRBox(e: any) {
    const text = e.currentTarget.dataset.text;
    if (!text) return;
    
    this.setData({ 
      showExplanation: true,
      explanation: { title: 'Thinking...', content: 'Asking AI for cultural details...', tags: [] }
    });
    
    try {
      const res = await wx.cloud.callFunction({
        name: 'aiService',
        data: {
          action: 'explain_menu_item',
          payload: {
            text,
            language: getLanguage()
          }
        }
      });
      
      const result = res.result as any;
      if (result.success) {
        const data = result.data;
        this.setData({
          explanation: {
            title: data.title || text,
            content: data.definition + (data.culture_note ? '\n\n' + data.culture_note : '') + (data.quote ? '\n\n"' + data.quote + '"' : ''),
            tags: data.tags || []
          }
        });
      }
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
      uploadedFileID: ''
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
      console.log('Save cancelled or failed', e);
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
        uploadedFileID: ''
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

    const langs = ['English', 'Chinese', 'Japanese', 'Korean', 'Spanish', 'French'];
    const langCodes = ['en', 'zh', 'ja', 'ko', 'es', 'fr'];

    try {
      const res = await wx.showActionSheet({ itemList: langs });
      const targetLang = langCodes[res.tapIndex];
      await this.startTranslation(targetLang);
    } catch (e) {
      console.log('Language selection cancelled');
    }
  },

  async startTranslation(targetLang: string) {
    this.setData({ translating: true, error: '' });
    wx.showLoading({ title: 'Translating...' });

    try {
      const fileID = await this.ensureImageUploaded();

      const callRes = await wx.cloud.callFunction({
        name: 'aiService',
        data: {
          action: 'translate_menu_image',
          payload: {
            fileID: fileID,
            source_lang: 'auto',
            target_lang: targetLang
          }
        }
      });

      const result = callRes.result as any;
      if (!result.success) throw new Error(result.message || 'Failed to start translation');
      const taskId = result.taskId;

      const maxRetries = 30; 
      let retries = 0;
      let translatedUrl = '';
      
      const getInterval = (retryCount: number) => {
        if (retryCount < 5) return 1000;
        if (retryCount < 10) return 2000;
        return 3000;
      };

      while (retries < maxRetries) {
        await new Promise(resolve => setTimeout(resolve, getInterval(retries)));
        
        const pollRes = await wx.cloud.callFunction({
          name: 'aiService',
          data: {
            action: 'get_translation_result',
            payload: { taskId }
          }
        });

        const pollResult = pollRes.result as any;
        if (pollResult.success && pollResult.status === 'SUCCEEDED') {
          translatedUrl = pollResult.imageUrl;
          break;
        } else if (pollResult.status === 'FAILED') {
          throw new Error(pollResult.message || 'Translation failed');
        }
        retries++;
      }

      if (translatedUrl) {
        this.setData({ translatedImageUrl: translatedUrl });
        
        // Show Preview
        wx.previewImage({ urls: [translatedUrl] });

        // ASK TO SAVE
        const modalRes = await wx.showModal({
            title: 'Translation Complete',
            content: 'Do you want to save the translated image to your album?',
            confirmText: 'Save',
            cancelText: 'Cancel'
        });

        if (modalRes.confirm) {
            await this.saveToAlbum(translatedUrl);
            wx.showToast({ title: 'Saved', icon: 'success' });
        }

      } else {
        throw new Error('Translation timed out');
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
