package com.travelgd.service;

import com.travelgd.common.ApiException;
import com.travelgd.dto.BlogCreateRequest;
import com.travelgd.mapper.TravelBlogMapper;
import com.travelgd.model.TravelBlog;
import com.travelgd.service.ai.QwenClient;
import org.springframework.stereotype.Service;

import java.util.List;
import java.util.Map;

@Service
public class BlogService {

    private final TravelBlogMapper blogMapper;
    private final QwenClient qwen;

    public BlogService(TravelBlogMapper blogMapper, QwenClient qwen) {
        this.blogMapper = blogMapper;
        this.qwen = qwen;
    }

    /** Public feed; likedByMe is only true when the viewer is authenticated and has liked it. */
    public List<TravelBlog> listFeed(Long viewerId) {
        return blogMapper.findAll(viewerId);
    }

    public TravelBlog getBlog(Long id, Long viewerId) {
        TravelBlog blog = blogMapper.findById(id, viewerId);
        if (blog == null) throw ApiException.notFound("Blog not found");
        return blog;
    }

    public TravelBlog publish(Long userId, BlogCreateRequest req) {
        TravelBlog blog = new TravelBlog();
        blog.setUserId(userId);
        blog.setTitle(req.title());
        blog.setContent(req.content());
        blog.setImages(req.images() == null ? List.of() : req.images());
        blogMapper.insert(blog);
        return getBlog(blog.getId(), userId);
    }

    public int like(Long id) {
        getBlog(id, null);
        blogMapper.incrementLikes(id);
        return getBlog(id, null).getLikes();
    }

    public Map<String, String> draft(Long userId, String topic) {
        String systemPrompt = """
                You are a travel blogger writing about Guangzhou and Lingnan culture. \
                Based on the user's topic or notes, produce a short blog post (200-300 words) \
                in English with a friendly, vivid tone. \
                Return ONLY valid JSON: { "title": string, "content": string }. Do NOT output markdown.""";
        String raw = qwen.chat("qwen-plus", List.of(
                Map.of("role", "system", "content", systemPrompt),
                Map.of("role", "user", "content", topic)));
        try {
            var node = new com.fasterxml.jackson.databind.ObjectMapper()
                    .readTree(QwenClient.stripFences(raw));
            return Map.of(
                    "title", node.path("title").asText(topic),
                    "content", node.path("content").asText(raw));
        } catch (Exception e) {
            return Map.of("title", topic, "content", raw);
        }
    }
}
