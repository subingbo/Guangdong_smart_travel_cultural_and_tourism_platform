import { getLanguage, getTranslations, setLanguage, Language } from '../../../utils/i18n';
import { fileUrl } from '../../../config/env';
import { getMe, listFavorites, listRoutes, myBlogs, updateMe, uploadAvatar } from '../../../services/api';
import { clearAuth, getToken, getUser } from '../../../utils/auth';
import { gotoLogin } from '../../../utils/request';

const DEFAULT_AVATAR_URL =
  'https://mmbiz.qpic.cn/mmbiz/icTdbqWNOwNRna42FI242Lcia07jQodd2FJGIYQfG0LAJGFxM4FbnQP6yfMxBgJ0F3YRqJCJ1aPAK2dQagdusBZg/0'

Page({
  data: {
    userInfo: {
      nickName: 'Guest User',
      avatarUrl: DEFAULT_AVATAR_URL,
    },
    userTag: 'Visitor',
    userId: '' as string | number,
    isLoggedIn: false,
    loading: false,
    avatarUploading: false,
    stats: {
      favorites: 0,
      history: 0,
      posts: 0,
      coupons: 0
    },
    favorites: [] as any[],
    blogs: [] as any[],
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
    this.refreshAuthState();
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
    // Re-render localized lists (names are language-dependent).
    if (this.data.isLoggedIn) {
      this.loadRemoteData();
    }
  },
  refreshAuthState() {
    const user = getUser();
    if (getToken() && user) {
      this.setData({
        isLoggedIn: true,
        userTag: 'Member',
        userId: user.id,
        userInfo: {
          nickName: user.nickname || 'Guest User',
          avatarUrl: user.avatarUrl ? fileUrl(user.avatarUrl) : DEFAULT_AVATAR_URL,
        }
      }, () => {
        this.loadRemoteData();
      });
    } else {
      this.setGuestState();
    }
  },
  async loadRemoteData() {
    if (!this.data.isLoggedIn) return;
    this.setData({ loading: true });
    const lang = getLanguage();
    try {
      const profile = await getMe();
      const [favorites, routes, blogsFeed] = await Promise.all([
        listFavorites(),
        listRoutes(),
        myBlogs(1, 10)
      ]);
      this.setData({
        userInfo: {
          nickName: profile.user.nickname || 'Guest User',
          avatarUrl: profile.user.avatarUrl ? fileUrl(profile.user.avatarUrl) : DEFAULT_AVATAR_URL,
        },
        stats: {
          favorites: (favorites || []).length,
          history: (routes || []).length,
          posts: profile.stats.posts,
          coupons: profile.stats.likesReceived
        },
        favorites: (favorites || []).map((p: any) => ({
          id: p.id,
          name: lang === 'zh' ? p.nameCn : p.nameEn,
          desc: lang === 'zh' ? p.descriptionCn : p.descriptionEn,
        })),
        blogs: (blogsFeed.items || []).map((b: any) => ({
          id: b.id,
          title: b.title,
          likes: b.likes,
          createdAt: (b.createdAt || '').substring(0, 10),
        })),
      });
    } catch (error) {
      console.error('Failed to load profile data', error);
    } finally {
      this.setData({ loading: false });
    }
  },
  goLogin() {
    gotoLogin();
  },
  onLogout() {
    clearAuth();
    this.setGuestState();
    wx.showToast({ title: '已退出', icon: 'success' });
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
      avatarUploading: false,
      stats: {
        favorites: 0,
        history: 0,
        posts: 0,
        coupons: 0
      },
      favorites: [],
      blogs: []
    });
  },
  onEditAvatar() {
    if (!this.data.isLoggedIn) {
      wx.showToast({ title: 'Please login first', icon: 'none' });
      return;
    }

    wx.chooseMedia({
      count: 1,
      mediaType: ['image'],
      sourceType: ['album', 'camera'],
      success: (res) => {
        const tempFilePath = res.tempFiles[0].tempFilePath;
        this.uploadAvatarFile(tempFilePath);
      }
    });
  },
  async uploadAvatarFile(filePath: string) {
    this.setData({ avatarUploading: true });
    wx.showLoading({ title: 'Uploading...' });

    try {
      // 1. Upload the file to the backend, 2. update the profile.
      const url = await uploadAvatar(filePath);
      await updateMe({ avatarUrl: url });
      this.setData({
        'userInfo.avatarUrl': fileUrl(url),
      });
      // Persist so other pages see the new avatar after relaunch.
      const storageUser = getUser();
      if (storageUser) {
        storageUser.avatarUrl = url;
        wx.setStorageSync('auth_user', storageUser);
      }
      wx.hideLoading();
      wx.showToast({ title: 'Updated successfully', icon: 'success' });
    } catch (error) {
      console.error('[uploadAvatar] failed', error);
      wx.hideLoading();
      wx.showToast({ title: 'Upload failed', icon: 'none' });
    } finally {
      this.setData({ avatarUploading: false });
    }
  },
  onTapGridItem(e: any) {
    const { index } = e.detail;
    // 0: Favorites, 1: History, 2: My Posts, 3: Likes received
    switch (index) {
      case 0:
        wx.showToast({
          title: this.data.isLoggedIn ? '我的收藏见下方列表' : '登录后可查看收藏',
          icon: 'none'
        });
        break;
      case 1:
        wx.navigateTo({ url: '/pages/user/routeInput/index' });
        break;
      case 2:
        wx.showToast({
          title: this.data.isLoggedIn ? '我的游记见下方列表' : '登录后可查看游记',
          icon: 'none'
        });
        break;
      case 3:
        wx.showToast({ title: '更多权益敬请期待', icon: 'none' });
        break;
    }
  },
  onTapHelp() {
    wx.showModal({
      title: this.data.t.profile.help || 'Help',
      content: 'For support, please contact support@example.com',
      showCancel: false
    });
  },
  onTapAbout() {
    wx.showModal({
      title: this.data.t.profile.about || 'About Us',
      content: 'Lingnan Travel AI Assistant v1.0.0\nDesigned for international visitors.',
      showCancel: false
    });
  }
})
