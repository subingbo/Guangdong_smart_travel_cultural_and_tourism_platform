const cloud = require('wx-server-sdk');
const axios = require('axios');

cloud.init({
  env: cloud.DYNAMIC_CURRENT_ENV
});

const API_KEY = process.env.DASHSCOPE_API_KEY;
// Use environment variable or fallback to the key provided by user
const TENCENT_MAP_API_KEY = process.env.TENCENT_MAP_API_KEY || '23ZBZ-6A46V-S2VPJ-5PSYS-2OQWH-QDF5L';

// Aliyun Bailian API Endpoints
const CHAT_URL = 'https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation/generation'; 
const WANX_ASYNC_URL = 'https://dashscope.aliyuncs.com/api/v1/services/aigc/image-generation/generation';
const WANX_TASK_URL = 'https://dashscope.aliyuncs.com/api/v1/tasks/';
const IMAGE_SYNTHESIS_URL = 'https://dashscope.aliyuncs.com/api/v1/services/aigc/image2image/image-synthesis';
const QWEN_OCR_URL = 'https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions';

// Outbound request guard: only http/https, and only known provider hosts.
// Rejects localhost, loopback, private and link-local addresses before any request.
const PRIVATE_HOST_RE = /^(localhost|127\.|0\.0\.0\.0|10\.|192\.168\.|169\.254\.|172\.(1[6-9]|2\d|3[01])\.|\[?::1\]?$)/i;

function assertSafeUrl(raw, allowHost) {
  const url = new URL(String(raw));
  if (url.protocol !== 'https:' && url.protocol !== 'http:') {
    throw new Error('Blocked URL protocol');
  }
  const host = url.hostname.toLowerCase();
  if (PRIVATE_HOST_RE.test(host) || host.endsWith('.local') || host.endsWith('.internal')) {
    throw new Error('Blocked private URL host');
  }
  if (!allowHost(host)) {
    throw new Error('Blocked URL host');
  }
  return url.toString();
}

const assertDashScopeUrl = (raw) =>
  assertSafeUrl(raw, (host) => host === 'dashscope.aliyuncs.com' || host.endsWith('.aliyuncs.com'));
const assertTencentUrl = (raw) => assertSafeUrl(raw, (host) => host === 'apis.map.qq.com');

const TASK_ID_RE = /^[A-Za-z0-9_-]{1,128}$/;
function requireTaskId(value) {
  if (typeof value !== 'string' || !TASK_ID_RE.test(value)) {
    throw new Error('Invalid taskId');
  }
  return value;
}

exports.main = async (event, context) => {
  const { action, payload } = event;
  const wxContext = cloud.getWXContext();
  const OPENID = wxContext.OPENID;

  if (!API_KEY) {
    return {
      success: false,
      message: 'DASHSCOPE_API_KEY is not set in environment variables'
    };
  }

  try {
    switch (action) {
      case 'chat':
        return await handleChat(payload);
      case 'generate_copy':
        return await handleCopy(payload);
      case 'generate_poster':
        return await handlePosterTaskCreate(payload);
      case 'get_poster_task':
        return await handlePosterTaskResult(payload, OPENID);
      case 'translate_menu_image':
        return await handleMenuTranslation(payload);
      case 'get_translation_result':
        return await handleTranslationResult(payload);
      case 'ocr_menu_image':
        return await handleMenuOCR(payload);
      case 'qwen_ocr': // New Qwen-OCR
        return await handleQwenOCR(payload);
      case 'extract_keywords': // New Keyword Extraction
        return await handleKeywordExtraction(payload);
      case 'generate_culture_card_start':
        return await handleCultureCardTaskCreate(payload, OPENID);
      case 'get_culture_card_result':
        return await handleCultureCardTaskResult(payload, OPENID);
      case 'generate_culture_card':
        return await handleCultureCardGeneration(payload, OPENID);
      case 'generate_culture_card_stream':
        return await handleCultureCardGenerationStream(payload, OPENID);
      case 'explain_menu_item':
        return await handleMenuExplanation(payload);
      case 'generate_smart_route':
        return await handleGenerateSmartRoute(payload, OPENID);
      case 'ai_guide_chat':
        return await handleAiGuideChat(payload);
      default:
        return {
          success: false,
          message: `Unknown action: ${action}`
        };
    }
  } catch (error) {
    console.error('AI Service Error:', error);
    return {
      success: false,
      message: error.message || 'Internal Server Error',
      details: error.response ? error.response.data : null
    };
  }
};

// ... (Existing functions: handleChat, handleCopy, handlePosterTaskCreate, handlePosterTaskResult)

async function handleChat(payload) {
  const { messages, model } = payload;
  const modelName = model || 'qwen-turbo'; 
  const requestBody = {
    model: modelName,
    input: { messages },
    parameters: { result_format: 'message' }
  };
  const response = await axios.post('https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation/generation', requestBody, {
    headers: { 'Authorization': `Bearer ${API_KEY}`, 'Content-Type': 'application/json' }
  });
  if (response.data.output && response.data.output.choices) {
    return { success: true, data: response.data.output.choices[0].message };
  } else {
    throw new Error('Invalid response from Chat API');
  }
}

async function handleCopy(payload) {
  const { text, languages, style } = payload;
  const modelName = 'qwen-turbo';
  const systemPrompt = `You are a professional social media copywriter. 
  Your task is to rewrite the input text into the specified style and translate it into the target languages.
  Input Text: "${text}"
  Style: ${style} (e.g., Instagram: visual/emoji-rich; Twitter: concise; Facebook: engaging story)
  Target Languages: ${languages.join(', ')}
  Return the result as a valid JSON object where keys are the language codes (${languages.join(', ')}) and values are the generated content. 
  Do not include any markdown formatting or extra text outside the JSON.`;
  const requestBody = {
    model: modelName,
    input: {
      messages: [
        { role: 'system', content: 'You are a helpful assistant that outputs JSON.' },
        { role: 'user', content: systemPrompt }
      ]
    },
    parameters: { result_format: 'message' }
  };
  const response = await axios.post('https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation/generation', requestBody, {
    headers: { 'Authorization': `Bearer ${API_KEY}`, 'Content-Type': 'application/json' }
  });
  if (response.data.output && response.data.output.choices) {
    const content = response.data.output.choices[0].message.content;
    const jsonStr = content.replace(/```json/g, '').replace(/```/g, '').trim();
    return { success: true, data: JSON.parse(jsonStr) };
  } else {
    throw new Error('Invalid response from Copy Generation API');
  }
}

async function handlePosterTaskCreate(payload) {
  const { prompt, size, n, style } = payload;
  const modelName = 'wan2.6-t2i';
  let finalPrompt = prompt;
  if (style && style !== 'Default') {
    finalPrompt += `, ${style} style`;
  }
  let finalSize = size || '1280*1280';
  if (!size) {
    const p = finalPrompt.toLowerCase();
    if (p.includes('vertical') || p.includes('portrait')) {
      finalSize = '960*1696'; 
    } else if (p.includes('horizontal') || p.includes('landscape')) {
      finalSize = '1696*960'; 
    }
  }
  const requestBody = {
    model: modelName,
    input: {
      messages: [{ role: 'user', content: [{ text: finalPrompt }] }]
    },
    parameters: { size: finalSize, n: n || 1, watermark: false, prompt_extend: true }
  };
  const response = await axios.post(assertDashScopeUrl(WANX_ASYNC_URL), requestBody, {
    headers: { 'Authorization': `Bearer ${API_KEY}`, 'Content-Type': 'application/json', 'X-DashScope-Async': 'enable' }
  });
  if (response.data.output && response.data.output.task_id) {
    return { success: true, taskId: response.data.output.task_id };
  }
  throw new Error('Invalid response from Poster Task Create API');
}

async function handlePosterTaskResult(payload, OPENID) {
  const { prompt, size, n, style } = payload;
  const taskId = requireTaskId(payload.taskId);
  const db = cloud.database();
  const existed = await db.collection('poster_generations').where({ taskId }).limit(1).get();
  if (existed.data && existed.data.length > 0) {
    const record = existed.data[0];
    if (record.imageUrl) {
      return { success: true, imageUrl: record.imageUrl, originalUrl: record.originalUrl || '', status: 'SUCCEEDED' };
    }
  }
  const response = await axios.get(assertDashScopeUrl(`${WANX_TASK_URL}${taskId}`), {
    headers: { 'Authorization': `Bearer ${API_KEY}` }
  });
  const output = response.data.output || {};
  const status = output.task_status;
  if (status !== 'SUCCEEDED') {
    return { success: true, status, pending: status === 'PENDING' || status === 'RUNNING' };
  }
  const choice = output.choices[0];
  const imageContent = choice.message && choice.message.content ? choice.message.content.find(c => c.image) : null;
  if (!imageContent) throw new Error('No image in task result');
  const imageUrl = imageContent.image;
  const imageRes = await axios.get(assertDashScopeUrl(imageUrl), { responseType: 'arraybuffer' });
  const buffer = imageRes.data;
  const timestamp = Date.now();
  const safeOpenId = OPENID || 'anonymous';
  const cloudPath = `images/${safeOpenId}/${timestamp}_${taskId}.png`;
  const uploadRes = await cloud.uploadFile({ cloudPath: cloudPath, fileContent: buffer });
  await savePosterRecord({
    openid: OPENID, prompt: prompt || '', imageUrl: uploadRes.fileID, originalUrl: imageUrl,
    size: size || '1280*1280', model: 'wan2.6-t2i', taskId, n: n || 1, style: style || ''
  });
  return { success: true, imageUrl: uploadRes.fileID, originalUrl: imageUrl, status: 'SUCCEEDED' };
}

async function savePosterRecord({ openid, prompt, imageUrl, originalUrl, size, model, taskId, n }) {
  const db = cloud.database();
  const serverDate = db.serverDate();
  let userInfo = {};
  if (openid) {
    const userRes = await db.collection('users').where({ openid }).limit(1).get();
    if (userRes.data && userRes.data.length > 0) {
      const user = userRes.data[0];
      userInfo = { userId: user._id || '', nickName: user.nickName || '', avatarUrl: user.avatarUrl || '' };
    }
  }
  await db.collection('poster_generations').add({
    data: {
      _openid: openid || '', openid: openid || '', prompt, imageUrl, originalUrl: originalUrl || '',
      size, model, taskId: taskId || '', imageCount: n || 1, ...userInfo, createdAt: serverDate, updatedAt: serverDate
    }
  });
}

// ... (Existing functions: handleMenuTranslation, handleMenuOCR - keep existing OCR as backup or rename)

async function handleMenuTranslation(payload) {
  const { image_url, fileID, source_lang, target_lang } = payload;
  const modelName = 'qwen-mt-image';
  let targetUrl = image_url;
  if (!targetUrl && fileID) {
    try {
      const res = await cloud.getTempFileURL({ fileList: [fileID] });
      if (res.fileList && res.fileList.length > 0) {
        targetUrl = res.fileList[0].tempFileURL;
      }
    } catch (e) {
      throw new Error('Failed to access the uploaded image');
    }
  }
  if (!targetUrl) throw new Error('image_url or fileID is required');
  const requestBody = {
    model: modelName,
    input: { image_url: targetUrl, source_lang: source_lang || 'auto', target_lang: target_lang || 'en' }
  };
  const response = await axios.post(assertDashScopeUrl(IMAGE_SYNTHESIS_URL), requestBody, {
    headers: { 'Authorization': `Bearer ${API_KEY}`, 'Content-Type': 'application/json', 'X-DashScope-Async': 'enable' }
  });
  if (response.data.output && response.data.output.task_id) {
    return { success: true, taskId: response.data.output.task_id };
  }
  throw new Error('Invalid response from Menu Translation API');
}

// Existing OCR using WeChat API (keep as 'ocr_menu_image')
async function handleMenuOCR(payload) {
  const { fileID, image_url } = payload;
  let imgUrl = image_url;
  if (!imgUrl && fileID) {
    const res = await cloud.getTempFileURL({ fileList: [fileID] });
    if (res.fileList && res.fileList.length > 0) {
      imgUrl = res.fileList[0].tempFileURL;
    }
  }
  if (!imgUrl) throw new Error('image_url or fileID is required for OCR');
  try {
    const result = await cloud.openapi.ocr.printedText({ type: 'photo', imgUrl: imgUrl });
    if (result.errCode === 0) {
      return {
        success: true,
        items: result.items.map(item => ({ text: item.text, box: item.pos }))
      };
    } else {
      throw new Error(`OCR Failed: ${result.errMsg}`);
    }
  } catch (e) {
    throw e;
  }
}

// NEW: Qwen-OCR
async function handleQwenOCR(payload) {
  const { fileID, image_url } = payload;
  let imgUrl = image_url;
  if (!imgUrl && fileID) {
    const res = await cloud.getTempFileURL({ fileList: [fileID] });
    if (res.fileList && res.fileList.length > 0) {
      imgUrl = res.fileList[0].tempFileURL;
    }
  }
  if (!imgUrl) throw new Error('image_url or fileID is required for Qwen-OCR');

  const requestBody = {
    model: "qwen-vl-ocr-latest",
    messages: [
      {
        role: "user",
        content: [
          { type: "image_url", image_url: { url: imgUrl } },
          { type: "text", text: "Please output only the text content from the image without any additional descriptions or formatting." }
        ]
      }
    ]
  };

  const response = await axios.post(assertDashScopeUrl(QWEN_OCR_URL), requestBody, {
    headers: { 'Authorization': `Bearer ${API_KEY}`, 'Content-Type': 'application/json' }
  });

  if (response.data.choices && response.data.choices.length > 0) {
    const content = response.data.choices[0].message.content;
    
    // Save to DB (optional but recommended by OCR.md)
    const db = cloud.database();
    await db.collection('ocr_records').add({
      data: {
        imageUrl: imgUrl,
        rawText: content,
        createdAt: db.serverDate()
      }
    });

    return { success: true, text: content };
  } else {
    throw new Error('Invalid response from Qwen-OCR API');
  }
}

// NEW: Keyword Extraction
async function handleKeywordExtraction(payload) {
  const { text } = payload;
  if (!text) throw new Error('Text is required for keyword extraction');

  const systemPrompt = `Analyze the following text (from a menu or scene). Extract ALL distinct text items, phrases, or menu names found.
  Do not omit any recognized item. If unsure about category, use 'Other'.
  Categorize each item into 'Food', 'Architecture', 'Custom', or 'Other' (default).
  Return a valid JSON object:
  {
    "keywords": [
      { "word": "Dim Sum", "category": "Food", "importance": 0.9 },
      { "word": "Cola", "category": "Other", "importance": 0.1 }
    ]
  }
  Only return JSON.`;

  const requestBody = {
    model: 'qwen-plus',
    input: {
      messages: [
        { role: 'system', content: 'You are a cultural expert assistant that outputs JSON.' },
        { role: 'user', content: systemPrompt + '\n\nText:\n' + text }
      ]
    },
    parameters: { result_format: 'message' }
  };

  const response = await axios.post('https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation/generation', requestBody, {
    headers: { 'Authorization': `Bearer ${API_KEY}`, 'Content-Type': 'application/json' }
  });

  if (response.data.output && response.data.output.choices) {
    const content = response.data.output.choices[0].message.content;
    const jsonStr = content.replace(/```json/g, '').replace(/```/g, '').trim();
    let data;
    try {
      data = JSON.parse(jsonStr);
    } catch (e) {
      const jsonMatch = jsonStr.match(/\{[\s\S]*\}/);
      if (jsonMatch) {
        data = JSON.parse(jsonMatch[0]);
      }
    }

    const normalizeKeywords = (items) => {
      if (!Array.isArray(items)) return [];
      return items
        .map((item) => ({
          word: String(item.word || item.text || '').trim(),
          category: item.category || 'Other',
          importance: typeof item.importance === 'number' ? item.importance : 0.1
        }))
        .filter((item) => item.word);
    };

    let keywords = normalizeKeywords(data && data.keywords);
    if (!keywords.length) {
      const tokens = text.match(/[A-Za-z0-9]+|[\u4e00-\u9fff]+/g) || [];
      const unique = Array.from(new Set(tokens));
      keywords = unique.map((word) => ({ word, category: 'Other', importance: 0.1 }));
    }
    data = { keywords };

    // Save to DB
    const db = cloud.database();
    await db.collection('keywords').add({
      data: {
        sourceTextSnippet: text.substring(0, 100),
        keywords: data.keywords,
        createdAt: db.serverDate()
      }
    });

    return { success: true, data };
  } else {
    throw new Error('Failed to extract keywords');
  }
}

// NEW: Culture Card Generation
async function handleCultureCardGeneration(payload, OPENID) {
  const { keyword } = payload;
  if (!keyword) throw new Error('Keyword is required');

  const systemPrompt = `Generate a South Guangdong Culture Card for the keyword: "${keyword}".
  Return a valid JSON object:
  {
    "title": "Title (e.g. Cantonese Morning Tea)",
    "description": "Short description (50 words)",
    "cultural_significance": "Cultural meaning (50 words)",
    "category": "Food/Architecture/Custom",
    "tags": ["Tag1", "Tag2"]
  }
  Only return JSON.`;

  const requestBody = {
    model: 'qwen-plus',
    input: {
      messages: [
        { role: 'system', content: 'You are a cultural expert assistant that outputs JSON.' },
        { role: 'user', content: systemPrompt }
      ]
    },
    parameters: { result_format: 'message', max_tokens: 512 }
  };

  const response = await axios.post('https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation/generation', requestBody, {
    headers: { 'Authorization': `Bearer ${API_KEY}`, 'Content-Type': 'application/json' }
  });

  if (response.data.output && response.data.output.choices) {
    const content = response.data.output.choices[0].message.content;
    const jsonStr = content.replace(/```json/g, '').replace(/```/g, '').trim();
    let data;
    try {
      data = JSON.parse(jsonStr);
    } catch (e) {
      const jsonMatch = jsonStr.match(/\{[\s\S]*\}/);
      if (jsonMatch) {
        data = JSON.parse(jsonMatch[0]);
      }
    }
    if (!data) throw new Error('Invalid JSON from culture card model');

    // Save to DB
    const db = cloud.database();
    await db.collection('culture_cards').add({
      data: {
        _openid: OPENID,
        keyword,
        cardData: data,
        createdAt: db.serverDate()
      }
    });

    return { success: true, data };
  } else {
    throw new Error('Failed to generate culture card');
  }
}

async function handleCultureCardGenerationStream(payload, OPENID) {
  const { keyword } = payload;
  if (!keyword) throw new Error('Keyword is required');

  const systemPrompt = `Generate a South Guangdong Culture Card for the keyword: "${keyword}".
  Return a valid JSON object:
  {
    "title": "Title (e.g. Cantonese Morning Tea)",
    "description": "Short description (50 words)",
    "cultural_significance": "Cultural meaning (50 words)",
    "category": "Food/Architecture/Custom",
    "tags": ["Tag1", "Tag2"]
  }
  Only return JSON.`;

  const requestBody = {
    model: 'qwen-plus',
    messages: [
      { role: 'system', content: 'You are a cultural expert assistant that outputs JSON.' },
      { role: 'user', content: systemPrompt }
    ],
    stream: true,
    max_tokens: 512,
    temperature: 0.01
  };

  const response = await axios.post(assertDashScopeUrl(QWEN_OCR_URL), requestBody, {
    headers: { 'Authorization': `Bearer ${API_KEY}`, 'Content-Type': 'application/json' },
    responseType: 'stream'
  });

  const streamResult = await new Promise((resolve, reject) => {
    let buffer = '';
    let content = '';
    const chunks = [];
    response.data.on('data', (chunk) => {
      buffer += chunk.toString('utf8');
      const lines = buffer.split('\n');
      buffer = lines.pop() || '';
      for (const line of lines) {
        const trimmed = line.trim();
        if (!trimmed || !trimmed.startsWith('data:')) continue;
        const dataStr = trimmed.replace(/^data:\s*/, '');
        if (dataStr === '[DONE]') continue;
        try {
          const json = JSON.parse(dataStr);
          const delta = json.choices && json.choices[0] && json.choices[0].delta;
          const piece = delta && delta.content ? delta.content : '';
          if (piece) {
            content += piece;
            chunks.push(piece);
          }
        } catch (e) {
        }
      }
    });
    response.data.on('end', () => resolve({ content, chunks }));
    response.data.on('error', reject);
  });

  const rawContent = streamResult.content || '';
  const cleaned = rawContent.replace(/```json/g, '').replace(/```/g, '').trim();
  let data;
  try {
    data = JSON.parse(cleaned);
  } catch (e) {
    const jsonMatch = cleaned.match(/\{[\s\S]*\}/);
    if (jsonMatch) {
      data = JSON.parse(jsonMatch[0]);
    }
  }

  if (data) {
    const db = cloud.database();
    await db.collection('culture_cards').add({
      data: {
        _openid: OPENID,
        keyword,
        cardData: data,
        createdAt: db.serverDate()
      }
    });
  }

  return {
    success: true,
    content: rawContent,
    chunks: streamResult.chunks,
    data: data || null
  };
}

async function handleCultureCardTaskCreate(payload, OPENID) {
  const { keyword } = payload;
  if (!keyword) throw new Error('Keyword is required');
  const db = cloud.database();
  const taskId = `cc_${Date.now()}_${Math.random().toString(36).slice(2, 8)}`;
  const now = db.serverDate();
  await db.collection('culture_cards').add({
    data: {
      _openid: OPENID,
      keyword,
      taskId,
      status: 'PENDING',
      createdAt: now,
      updatedAt: now
    }
  });
  return { success: true, taskId };
}

async function handleCultureCardTaskResult(payload, OPENID) {
  const rawTaskId = payload.taskId;
  if (!rawTaskId) throw new Error('taskId is required');
  const taskId = requireTaskId(rawTaskId);
  if (!/^[A-Za-z0-9_-]{1,128}$/.test(taskId)) throw new Error('Invalid taskId');
  const db = cloud.database();
  const _ = db.command;
  const res = await db.collection('culture_cards').where({ taskId }).limit(1).get();
  if (!res.data || res.data.length === 0) {
    return { success: false, status: 'NOT_FOUND', message: 'Task not found' };
  }
  const task = res.data[0];
  if (task.status === 'SUCCEEDED') {
    return { success: true, status: 'SUCCEEDED', data: task.cardData };
  }
  if (task.status === 'FAILED') {
    return { success: false, status: 'FAILED', message: task.error || 'Generation failed' };
  }
  const lock = await db.collection('culture_cards').where({ taskId, status: 'PENDING' }).update({
    data: { status: 'RUNNING', updatedAt: db.serverDate(), startedAt: db.serverDate() }
  });
  if (lock.stats && lock.stats.updated === 0) {
    return { success: true, status: 'RUNNING' };
  }
  try {
    const result = await handleCultureCardGeneration({ keyword: task.keyword }, OPENID);
    await db.collection('culture_cards').where({ taskId }).update({
      data: { status: 'SUCCEEDED', cardData: result.data, updatedAt: db.serverDate() }
    });
    return { success: true, status: 'SUCCEEDED', data: result.data };
  } catch (e) {
    await db.collection('culture_cards').where({ taskId }).update({
      data: { status: 'FAILED', error: e.message || 'Generation failed', updatedAt: db.serverDate() }
    });
    return { success: false, status: 'FAILED', message: e.message || 'Generation failed' };
  }
}

// ... (Existing functions: handleMenuExplanation, handleTranslationResult, handleGenerateSmartRoute, handleAiGuideChat)

async function handleMenuExplanation(payload) {
  const { text, language } = payload;
  const targetLang = language || 'zh'; 
  const db = cloud.database();
  const _ = db.command;
  const collection = db.collection('word_content');
  const cacheKey = text.trim().toLowerCase();
  
  try {
    const cacheRes = await collection.where({ word: cacheKey, lang: targetLang }).get();
    if (cacheRes.data.length > 0) {
      await collection.doc(cacheRes.data[0]._id).update({ data: { view_count: _.inc(1) } });
      return { success: true, data: cacheRes.data[0].data };
    }
  } catch (e) { }

  const systemPrompt = `You are a culinary expert and cultural ambassador. Explain the following menu item in a concise, engaging way. 
  Output MUST be a valid JSON object with fields: title, definition, culture_note, quote, tags.
  Ensure content is in ${targetLang === 'en' ? 'English' : 'Chinese'}. Do NOT output markdown.`;

  const requestBody = {
    model: 'qwen-max', 
    input: {
      messages: [
        { role: 'system', content: systemPrompt },
        { role: 'user', content: text }
      ]
    }
  };

  try {
    const response = await axios.post('https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation/generation', requestBody, {
      headers: { 'Authorization': `Bearer ${API_KEY}`, 'Content-Type': 'application/json' }
    });

    if (response.data.output && response.data.output.choices) {
      let content = response.data.output.choices[0].message.content;
      content = content.replace(/```json/g, '').replace(/```/g, '').trim();
      let parsedData;
      try { parsedData = JSON.parse(content); } catch (e) {
        parsedData = { title: text, definition: content, culture_note: '', quote: '', tags: [] };
      }
      try {
        await collection.add({ data: { word: cacheKey, lang: targetLang, data: parsedData, view_count: 1, createdAt: db.serverDate() } });
      } catch (e) { }
      return { success: true, data: parsedData };
    } else {
      throw new Error('Invalid response from Explanation API');
    }
  } catch (error) { throw error; }
}

async function handleTranslationResult(payload) {
  const rawTaskId = payload.taskId;
  if (!rawTaskId) throw new Error('taskId is required');
  const taskId = requireTaskId(rawTaskId);
  const taskUrl = new URL(`${WANX_TASK_URL}${encodeURIComponent(taskId)}`);
  if (taskUrl.protocol !== 'https:' || taskUrl.hostname !== 'dashscope.aliyuncs.com') {
    throw new Error('Blocked task URL');
  }
  const response = await axios.get(taskUrl.toString(), {
    headers: { 'Authorization': `Bearer ${API_KEY}` },
    maxRedirects: 0,
  });
  const output = response.data.output || {};
  const status = output.task_status;
  if (status !== 'SUCCEEDED') {
    return { success: true, status, pending: status === 'PENDING' || status === 'RUNNING', message: output.message || '' };
  }
  const imageUrl = output.image_url;
  if (!imageUrl) throw new Error('No image_url in task result');
  return { success: true, status: 'SUCCEEDED', imageUrl: imageUrl };
}

async function handleGenerateSmartRoute(payload, OPENID) {
  const { duration, interests, pace, mode = 'driving', location = "Guangzhou" } = payload;
  const modelName = 'qwen-plus'; 
  
  if (!TENCENT_MAP_API_KEY) {
    console.warn('TENCENT_MAP_API_KEY is not configured. Route polyline will not be generated.');
  }

  const systemPrompt = `You are an expert travel planner for ${location}. Generate a detailed ${duration}-day itinerary with a ${pace} pace, focusing on these interests: ${interests.join(', ')}.
  You MUST return a VALID JSON object matching the structure: { routeName, duration, totalDistance, waypoints: [{day, name, latitude, longitude, description, stayTime}] }.
  Ensure coordinates are accurate for ${location}.`;
  const requestBody = {
    model: modelName,
    input: {
      messages: [
        { role: 'system', content: 'You are a helpful assistant that outputs strict JSON.' },
        { role: 'user', content: systemPrompt }
      ]
    },
    parameters: { result_format: 'message' }
  };
  try {
    const response = await axios.post('https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation/generation', requestBody, { headers: { 'Authorization': `Bearer ${API_KEY}`, 'Content-Type': 'application/json' } });
    if (response.data.output && response.data.output.choices) {
      let content = response.data.output.choices[0].message.content;
      content = content.replace(/```json/g, '').replace(/```/g, '').trim();
      let routeData = JSON.parse(content);
      const tencentRoute = await fetchTencentRoutePolyline(routeData && routeData.waypoints ? routeData.waypoints : [], mode);
      if (tencentRoute && tencentRoute.points.length > 1) {
        routeData.polyline = tencentRoute.points.map(p => `${p.latitude},${p.longitude}`);
        if (tencentRoute.distance) {
          routeData.totalDistance = formatDistance(tencentRoute.distance);
        }
      } else {
        // Fallback: Generate straight lines between waypoints if API fails
        console.warn('Using fallback straight-line polyline due to missing Map API key or error.');
        const wps = routeData.waypoints || [];
        if (wps.length >= 2) {
           routeData.polyline = wps.map(wp => `${wp.latitude},${wp.longitude}`);
        }
      }
      const db = cloud.database();
      const res = await db.collection('routes').add({
        data: { _openid: OPENID, ...routeData, createdAt: db.serverDate(), originalRequest: payload }
      });
      return { success: true, routeId: res._id, data: routeData };
    } else { throw new Error('Invalid response from Route Gen API'); }
  } catch (error) { throw error; }
}

function formatDistance(distance) {
  if (!distance || typeof distance !== 'number') return '';
  if (distance >= 1000) {
    return `${(distance / 1000).toFixed(1)}km`;
  }
  return `${Math.round(distance)}m`;
}

function decodeTencentPolyline(polyline) {
  if (!Array.isArray(polyline) || polyline.length < 2) return [];
  const pl = polyline.slice();
  
  // Check if the first point is encoded (multiplied by 1e6)
  // Latitude is [-90, 90], so if abs > 90 (safely 180), it's encoded.
  if (Math.abs(pl[0]) > 180) {
    pl[0] = pl[0] / 1000000;
    pl[1] = pl[1] / 1000000;
  }
  
  // The rest are always deltas * 1e6
  for (let i = 2; i < pl.length; i++) {
    pl[i] = Number((pl[i - 2] + pl[i] / 1000000).toFixed(6));
  }
  
  const points = [];
  for (let i = 0; i < pl.length; i += 2) {
    points.push({ latitude: pl[i], longitude: pl[i + 1] });
  }
  // Debug log for first few points to verify decoding
  if (points.length > 0) {
    console.log('Decoded Polyline Sample:', points.slice(0, 3));
  }
  return points;
}

function splitRouteSegments(points, maxWaypoints = 10) {
  const segments = [];
  if (points.length <= 2) return [points];
  let startIndex = 0;
  while (startIndex < points.length - 1) {
    const remaining = points.length - startIndex;
    const segmentPoints = Math.min(remaining, maxWaypoints + 2);
    const endIndex = startIndex + segmentPoints - 1;
    segments.push(points.slice(startIndex, endIndex + 1));
    startIndex = endIndex;
  }
  return segments;
}

async function requestTencentRoute(points, mode = 'driving') {
  if (!TENCENT_MAP_API_KEY || points.length < 2) {
    console.warn('Skipping Tencent Map Request: Missing Key or Insufficient Points', { hasKey: !!TENCENT_MAP_API_KEY, points: points.length });
    return null;
  }
  const from = `${points[0].latitude},${points[0].longitude}`;
  const to = `${points[points.length - 1].latitude},${points[points.length - 1].longitude}`;
  
  // Waypoints limit is usually 10 for driving API. Ensure we don't exceed if logic upstream fails.
  const waypointsList = points.slice(1, -1);
  const waypoints = waypointsList.map(p => `${p.latitude},${p.longitude}`).join(';');
  
  const params = {
    from,
    to,
    key: TENCENT_MAP_API_KEY
  };
  if (waypoints) {
    params.waypoints = waypoints;
  }
  
  try {
    console.log('Requesting Tencent Map Direction:', { from, to, waypointsCount: waypointsList.length, mode });
    const response = await axios.get(assertTencentUrl(`https://apis.map.qq.com/ws/direction/v1/${mode}/`), { params });
    
    if (!response.data || response.data.status !== 0) {
      console.error('Tencent Map API Error:', response.data);
      return null;
    }
    
    if (!response.data.result || !response.data.result.routes || response.data.result.routes.length === 0) {
      console.warn('Tencent Map API returned no routes');
      return null;
    }

    const route = response.data.result.routes[0];
    let pointsDecoded = [];
    if (mode === 'transit') {
      if (route.steps) {
        for (const step of route.steps) {
          if (step.polyline) {
            pointsDecoded = pointsDecoded.concat(decodeTencentPolyline(step.polyline));
          } else if (step.lines && step.lines.length > 0 && step.lines[0].polyline) {
             pointsDecoded = pointsDecoded.concat(decodeTencentPolyline(step.lines[0].polyline));
          }
        }
      }
    } else {
      pointsDecoded = decodeTencentPolyline(route.polyline || []);
    }
    
    return {
      points: pointsDecoded,
      distance: route.distance || 0,
      duration: route.duration || 0
    };
  } catch (error) {
    console.error('Tencent Map Request Failed:', error.message, error.response ? error.response.data : '');
    return null;
  }
}

async function fetchTencentRoutePolyline(waypoints, mode = 'driving') {
  if (!TENCENT_MAP_API_KEY || !Array.isArray(waypoints) || waypoints.length < 2) return null;
  const basePoints = waypoints
    .map(wp => ({ latitude: Number(wp.latitude), longitude: Number(wp.longitude) }))
    .filter(p => Number.isFinite(p.latitude) && Number.isFinite(p.longitude));
  if (basePoints.length < 2) return null;
  const maxWaypoints = (mode === 'transit') ? 0 : 10;
  const segments = splitRouteSegments(basePoints, maxWaypoints);
  let allPoints = [];
  let totalDistance = 0;
  let totalDuration = 0;
  for (const segment of segments) {
    const res = await requestTencentRoute(segment, mode);
    if (!res || res.points.length === 0) continue;
    totalDistance += res.distance || 0;
    totalDuration += res.duration || 0;
    const segmentPoints = allPoints.length > 0 ? res.points.slice(1) : res.points;
    allPoints = allPoints.concat(segmentPoints);
  }
  return {
    points: allPoints,
    distance: totalDistance,
    duration: totalDuration
  };
}

async function handleAiGuideChat(payload) {
  const { question, history = [], context, type } = payload;
  const modelName = 'qwen-turbo';
  
  let systemPrompt = `You are a knowledgeable AI tour guide. Answer the user's question about travel, history, or culture. Keep answers concise (under 100 words) and engaging.`;
  
  if (context) {
    systemPrompt += ` Current Location Context: ${JSON.stringify(context)}.`;
  }

  let messages = [];
  if (type === 'welcome') {
    systemPrompt += ` Generate a warm, short welcome message based on the user's location. Briefly mention a nearby landmark if known from coordinates, or just welcome them to the area. Introduce yourself.`;
    messages = [{ role: 'system', content: systemPrompt }, { role: 'user', content: 'Hello, where am I?' }];
  } else {
    messages = [{ role: 'system', content: systemPrompt }, ...history, { role: 'user', content: question }];
  }

  const requestBody = {
    model: modelName,
    input: { messages: messages },
    parameters: { result_format: 'message' }
  };
  try {
    const response = await axios.post('https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation/generation', requestBody, { headers: { 'Authorization': `Bearer ${API_KEY}`, 'Content-Type': 'application/json' } });
    if (response.data.output && response.data.output.choices) {
      return { success: true, data: response.data.output.choices[0].message };
    } else { throw new Error('Invalid response from Chat API'); }
  } catch (error) { throw error; }
}
