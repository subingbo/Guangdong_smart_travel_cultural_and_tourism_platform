import { getLanguage, getTranslations } from '../../../utils/i18n';
import { Blog, draftBlog, likeBlog, listBlogs, publishBlog, unlikeBlog, uploadBlogImage } from '../../../services/api';
import { getToken } from '../../../utils/auth';
import { gotoLogin } from '../../../utils/request';
import { fileUrl } from '../../../config/env';

Page({
  data: {
    t: {} as any,
    lang: 'en',
    tempImages: [] as string[],
    topic: '',
    loading: false,
    publishing: false,
    generatedBlog: null as { title: string; content: string } | null,
    blogs: [] as Blog[],
  },

  onShow() {
    if (typeof this.getTabBar === 'function' && this.getTabBar()) {
      this.getTabBar().setData({
        selected: 'pages/user/travelBlog/index'
      })
    }
    const lang = getLanguage();
    this.setData({ t: getTranslations(lang), lang });
    wx.setNavigationBarTitle({
      title: lang === 'en' ? 'Travel Blog' : '智慧游记'
    });
    this.refreshFeed();
  },

  async refreshFeed() {
    try {
      const feed = await listBlogs(1, 10);
      const blogs = (feed.items || []).map((b: Blog) => ({
        ...b,
        coverImage: b.images && b.images.length > 0 ? fileUrl(b.images[0]) : '',
        authorName: b.author ? b.author.nickname : 'Traveler',
        authorAvatar: b.author && b.author.avatarUrl ? fileUrl(b.author.avatarUrl) : '',
      }));
      this.setData({ blogs });
    } catch (err) {
      console.error('Failed to load blog feed:', err);
    }
  },

  onChooseImage() {
    const remain = 3 - this.data.tempImages.length;
    if (remain <= 0) {
      wx.showToast({ title: '最多 3 张图片', icon: 'none' });
      return;
    }
    wx.chooseMedia({
      count: remain,
      mediaType: ['image'],
      sourceType: ['album', 'camera'],
      success: (res) => {
        const paths = res.tempFiles.map((f) => f.tempFilePath);
        this.setData({ tempImages: this.data.tempImages.concat(paths) });
      }
    });
  },

  onRemoveImage(e: any) {
    const index = e.currentTarget.dataset.index;
    const tempImages = this.data.tempImages.filter((_, i) => i !== index);
    this.setData({ tempImages });
  },

  onTopic(e: any) {
    this.setData({ topic: e.detail.value });
  },

  onTitleInput(e: any) {
    if (this.data.generatedBlog) {
      this.setData({ 'generatedBlog.title': e.detail.value });
    }
  },

  onContentInput(e: any) {
    if (this.data.generatedBlog) {
      this.setData({ 'generatedBlog.content': e.detail.value });
    }
  },

  requireLogin(): boolean {
    if (getToken()) return true;
    wx.showModal({
      title: '需要登录',
      content: 'AI 起草与发布游记需要先登录，是否前往登录页？',
      confirmText: '去登录',
      success: (res) => {
        if (res.confirm) gotoLogin();
      }
    });
    return false;
  },

  async onGenerate() {
    const topic = this.data.topic.trim();
    if (!topic) {
      wx.showToast({ title: '先写一句主题，例如：夜游珠江', icon: 'none' });
      return;
    }
    if (!this.requireLogin()) return;

    this.setData({ loading: true });
    try {
      const draft = await draftBlog(topic);
      this.setData({
        generatedBlog: { title: draft.title, content: draft.content }
      });
      wx.showToast({ title: 'AI 已起草', icon: 'success' });
    } catch (err) {
      console.error('Blog draft failed:', err);
    } finally {
      this.setData({ loading: false });
    }
  },

  async onPublish() {
    const blog = this.data.generatedBlog;
    if (!blog) return;
    if (!blog.title.trim() || !blog.content.trim()) {
      wx.showToast({ title: '标题和内容不能为空', icon: 'none' });
      return;
    }
    if (!this.requireLogin()) return;

    this.setData({ publishing: true });
    try {
      // Upload the local images first, then publish with their backend URLs.
      const urls: string[] = [];
      for (const path of this.data.tempImages) {
        urls.push(await uploadBlogImage(path));
      }
      await publishBlog(blog.title.trim(), blog.content.trim(), urls);
      wx.showToast({ title: '已发布', icon: 'success' });
      this.setData({
        generatedBlog: null,
        tempImages: [],
        topic: ''
      });
      this.refreshFeed();
    } catch (err) {
      console.error('Publish failed:', err);
    } finally {
      this.setData({ publishing: false });
    }
  },

  async onLike(e: any) {
    if (!getToken()) {
      gotoLogin();
      return;
    }
    const index = e.currentTarget.dataset.index;
    const blog = this.data.blogs[index];
    if (!blog) return;
    try {
      if (blog.likedByMe) {
        await unlikeBlog(blog.id);
        blog.likes -= 1;
        blog.likedByMe = false;
      } else {
        await likeBlog(blog.id);
        blog.likes += 1;
        blog.likedByMe = true;
      }
      this.setData({ [`blogs[${index}]`]: blog });
    } catch (err) {
      console.error('Like failed:', err);
    }
  },

  onCancelEdit() {
    this.setData({ generatedBlog: null });
  }
})
