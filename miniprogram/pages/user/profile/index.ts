import { getLanguage, getTranslations, setLanguage, Language } from '../../../utils/i18n';

const DEFAULT_AVATAR_URL =
  'https://mmbiz.qpic.cn/mmbiz/icTdbqWNOwNRna42FI242Lcia07jQodd2FJGIYQfG0LAJGFxM4FbnQP6yfMxBgJ0F3YRqJCJ1aPAK2dQagdusBZg/0'
const STORAGE_KEY = 'userProfile'

Page({
  data: {
    userInfo: {
      nickName: 'Guest User',
      avatarUrl: DEFAULT_AVATAR_URL,
    },
    userTag: 'Visitor',
    userId: '',
    openid: '',
    isLoggedIn: false,
    loading: false,
    posterLoading: false,
    posterHistory: [] as any[],
    stats: {
      favorites: 0,
      history: 0,
      posts: 0,
      coupons: 0
    },
    t: {} as any, // Translations
    currentLang: 'en' as Language
  },
  onShow() {
    if (typeof this.getTabBar === 'function' && this.getTabBar()) {
      this.getTabBar().setData({
        selected: 'pages/user/profile/index'
      })
    }
    this.initLanguage();
    this.loadUserFromCache()
    this.loadUserFromCloud()
  },
  initLanguage() {
    const lang = getLanguage();
    const t = getTranslations(lang);
    this.setData({
      currentLang: lang,
      t: t
    });
    wx.setNavigationBarTitle({
      title: lang === 'en' ? 'Profile' : '个人中心'
    });
    
    if (typeof this.getTabBar === 'function' && this.getTabBar()) {
      this.getTabBar().setData({
        t: t
      });
    }
  },
  onTapLanguage() {
    const newLang = this.data.currentLang === 'en' ? 'zh' : 'en';
    setLanguage(newLang);
    this.initLanguage();
    wx.showToast({
      title: newLang === 'en' ? 'Switched to English' : '已切换至中文',
      icon: 'none'
    });
    // Optional: Re-launch to update TabBar and other pages if they don't listen to onShow
    // wx.reLaunch({ url: '/pages/user/profile/index' });
  },
  async loadStats() {
    if (!this.data.isLoggedIn) return
    const db = wx.cloud.database()
    try {
      const [favRes, routeRes, blogRes] = await Promise.all([
        db.collection('favorites').count(),
        db.collection('user_routes').count(),
        db.collection('travel_blogs').count()
      ])
      
      this.setData({
        'stats.favorites': favRes.total,
        'stats.history': routeRes.total,
        'stats.posts': blogRes.total
      })
    } catch (error) {
      console.error('Failed to load stats', error)
    }
  },
  loadUserFromCache() {
    const cachedUser = wx.getStorageSync(STORAGE_KEY)
    if (cachedUser && cachedUser.nickName) {
      this.setData({
        userInfo: cachedUser,
        isLoggedIn: true,
        userTag: 'Member',
      }, () => {
        this.loadStats()
      })
    }
  },
  async loadUserFromCloud() {
    if (!wx.cloud) {
      return
    }
    this.setData({ loading: true })
    try {
      const loginRes = await wx.cloud.callFunction({ name: 'login' })
      const { openid } = loginRes.result as { openid: string }
      if (!openid) {
        throw new Error('missing openid')
      }
      const db = wx.cloud.database()
      const userRes = await db.collection('users').where({ openid }).limit(1).get()
      if (userRes.data && userRes.data.length > 0) {
        const user = userRes.data[0] as { _id?: string; nickName?: string; avatarUrl?: string }
        const userInfo = {
          nickName: user.nickName || 'Guest User',
          avatarUrl: user.avatarUrl || DEFAULT_AVATAR_URL,
        }
        this.setData({
          userInfo,
          isLoggedIn: true,
          userTag: 'Member',
          userId: user._id || '',
          openid,
        }, () => {
          this.loadStats()
          this.loadPosterHistory()
        })
        wx.setStorageSync(STORAGE_KEY, userInfo)
        return
      }
      this.setGuestState()
    } catch (error) {
      console.error('[loadUserFromCloud] 失败', error)
      this.setGuestState()
      wx.showToast({ title: '加载用户失败，请检查控制台', icon: 'none' })
    } finally {
      this.setData({ loading: false })
    }
  },
  onRegister() {
    if (!wx.getUserProfile) {
      wx.showToast({ title: '当前基础库不支持', icon: 'none' })
      return
    }
    this.setData({ loading: true })
    wx.getUserProfile({
      desc: '用于完善个人资料',
      success: async (res) => {
        try {
          const loginRes = await wx.cloud.callFunction({ name: 'login' })
          const { openid } = loginRes.result as { openid: string }
          if (!openid) {
            throw new Error('missing openid')
          }
          const nickName = res.userInfo?.nickName || 'Guest User'
          const avatarUrl = res.userInfo?.avatarUrl || DEFAULT_AVATAR_URL
          const db = wx.cloud.database()
          const serverDate = db.serverDate()
          const existed = await db.collection('users').where({ openid }).limit(1).get()
          let userId = ''
          if (existed.data && existed.data.length > 0) {
            const docId = existed.data[0]._id as string
            userId = docId
            await db.collection('users').doc(docId).update({
              data: {
                nickName,
                avatarUrl,
                updatedAt: serverDate,
              },
            })
          } else {
            const addRes = await db.collection('users').add({
              data: {
                openid,
                nickName,
                avatarUrl,
                createdAt: serverDate,
                updatedAt: serverDate,
              },
            })
            userId = addRes._id as string
          }
          const userInfo = { nickName, avatarUrl }
          this.setData({
            userInfo,
            isLoggedIn: true,
            userTag: 'Member',
            userId,
            openid,
          }, () => {
            this.loadStats()
            this.loadPosterHistory()
          })
          wx.setStorageSync(STORAGE_KEY, userInfo)
          wx.showToast({ title: '登录成功', icon: 'success' })
        } catch (error) {
          console.error('[onRegister] 失败', error)
          this.setGuestState()
          wx.showToast({ title: '登录失败，请检查控制台', icon: 'none' })
        } finally {
          this.setData({ loading: false })
        }
      },
      fail: () => {
        this.setData({ loading: false })
      },
    })
  },
  onLogout() {
    wx.removeStorageSync(STORAGE_KEY)
    this.setGuestState()
    wx.showToast({ title: '已退出', icon: 'success' })
  },
  setGuestState() {
    this.setData({
      userInfo: {
        nickName: 'Guest User',
        avatarUrl: DEFAULT_AVATAR_URL,
      },
      isLoggedIn: false,
      userTag: 'Visitor',
      userId: '',
      openid: '',
      posterHistory: [],
      posterLoading: false,
      stats: {
        favorites: 0,
        history: 0,
        posts: 0,
        coupons: 0
      }
    })
  },
  async loadPosterHistory() {
    if (!this.data.openid) return
    this.setData({ posterLoading: true })
    const db = wx.cloud.database()
    try {
      const res = await db.collection('poster_generations')
        .where({ openid: this.data.openid })
        .orderBy('createdAt', 'desc')
        .limit(20)
        .get()
      const list = (res.data || []).map((item: any, index: number) => {
        const createdAt = item.createdAt ? new Date(item.createdAt).toLocaleString() : ''
        return {
          ...item,
          displayTime: createdAt,
          id: item._id || `${index}`
        }
      })
      this.setData({ posterHistory: list })
    } catch (error) {
      console.error('Failed to load poster history', error)
      this.setData({ posterHistory: [] })
    } finally {
      this.setData({ posterLoading: false })
    }
  },
  onPreviewPosterHistory(e: any) {
    const url = e.currentTarget.dataset.url
    if (url) {
      wx.previewImage({
        urls: [url],
        current: url
      })
    }
  },
  onEditAvatar() {
    if (!this.data.isLoggedIn) {
      wx.showToast({ title: 'Please login first', icon: 'none' })
      return
    }
    
    wx.chooseMedia({
      count: 1,
      mediaType: ['image'],
      sourceType: ['album', 'camera'],
      success: (res) => {
        const tempFilePath = res.tempFiles[0].tempFilePath
        this.uploadAvatar(tempFilePath)
      }
    })
  },
  async uploadAvatar(filePath: string) {
    if (!this.data.userId) {
      wx.showToast({ title: 'User ID missing', icon: 'none' })
      return
    }
    
    wx.showLoading({ title: 'Uploading...' })
    
    try {
      const ext = filePath.match(/\.[^.]+$/)?.[0] || '.png'
      const cloudPath = `avatars/${this.data.userId}${ext}`
      
      const uploadRes = await wx.cloud.uploadFile({
        cloudPath,
        filePath,
      })
      
      const fileID = uploadRes.fileID
      
      // Update DB
      const db = wx.cloud.database()
      await db.collection('users').doc(this.data.userId).update({
        data: {
          avatarUrl: fileID,
          updatedAt: db.serverDate()
        }
      })
      
      // Update local state
      this.setData({
        'userInfo.avatarUrl': fileID
      })
      
      // Update cache
      const cachedUser = wx.getStorageSync(STORAGE_KEY) || {}
      cachedUser.avatarUrl = fileID
      wx.setStorageSync(STORAGE_KEY, cachedUser)
      
      wx.hideLoading()
      wx.showToast({ title: 'Updated successfully', icon: 'success' })
      
    } catch (error) {
      console.error('[uploadAvatar] failed', error)
      wx.hideLoading()
      wx.showToast({ title: 'Upload failed', icon: 'none' })
    }
  },
  onTapGridItem(e: any) {
    const { index } = e.detail
    // 0: Favorites, 1: History, 2: My Posts, 3: Coupons
    switch (index) {
      case 0:
        wx.showToast({ title: 'Favorites list coming soon', icon: 'none' })
        break
      case 1:
        wx.navigateTo({ url: '/pages/user/routeInput/index' })
        break
      case 2:
        wx.showToast({ title: 'My Posts list coming soon', icon: 'none' })
        break
      case 3:
        wx.showToast({ title: 'Coupons feature coming soon', icon: 'none' })
        break
    }
  },
  // onTapLanguage() removed as it is now handled in initLanguage and onTapLanguage above
  onTapHelp() {
    wx.showModal({
      title: this.data.t.profile.help || 'Help',
      content: 'For support, please contact support@example.com',
      showCancel: false
    })
  },
  onTapAbout() {
    wx.showModal({
      title: this.data.t.profile.about || 'About Us',
      content: 'Lingnan Travel AI Assistant v1.0.0\nDesigned for international visitors.',
      showCancel: false
    })
  },
  onInitDB() {
    wx.showLoading({ title: '初始化中...' })
    wx.cloud.callFunction({
      name: 'initDB',
      success: (res: any) => {
        console.log('[initDB] success', res)
        wx.hideLoading()
        wx.showModal({
          title: '数据库初始化结果',
          content: JSON.stringify(res.result.results, null, 2),
          showCancel: false
        })
      },
      fail: (err) => {
        console.error('[initDB] fail', err)
        wx.hideLoading()
        wx.showToast({ title: '初始化失败', icon: 'none' })
      }
    })
  },
  onSeedData() {
    wx.showLoading({ title: '导入中...' })
    wx.cloud.callFunction({
      name: 'seedData',
      success: (res: any) => {
        console.log('[seedData] success', res)
        wx.hideLoading()
        wx.showModal({
          title: '导入完成',
          content: '已为各集合导入 5 条示例数据',
          showCancel: false
        })
      },
      fail: (err) => {
        console.error('[seedData] fail', err)
        wx.hideLoading()
        wx.showToast({ title: '导入失败', icon: 'none' })
      }
    })
  }
})
