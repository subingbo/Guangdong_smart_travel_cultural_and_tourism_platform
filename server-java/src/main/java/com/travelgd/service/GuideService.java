package com.travelgd.service;

import com.travelgd.common.ApiException;
import com.travelgd.dto.GuideChatRequest;
import com.travelgd.mapper.GuideMessageMapper;
import com.travelgd.mapper.GuideSessionMapper;
import com.travelgd.mapper.KnowledgeMapper;
import com.travelgd.model.GuideSession;
import com.travelgd.model.Knowledge;
import com.travelgd.service.ai.QwenClient;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;

/**
 * AI tour guide with a lightweight RAG: before calling the model, keyword hits
 * from the knowledge table are stitched into the system prompt.
 */
@Service
public class GuideService {

    private final GuideSessionMapper sessionMapper;
    private final GuideMessageMapper messageMapper;
    private final KnowledgeMapper knowledgeMapper;
    private final QwenClient qwen;

    public GuideService(GuideSessionMapper sessionMapper, GuideMessageMapper messageMapper,
                        KnowledgeMapper knowledgeMapper, QwenClient qwen) {
        this.sessionMapper = sessionMapper;
        this.messageMapper = messageMapper;
        this.knowledgeMapper = knowledgeMapper;
        this.qwen = qwen;
    }

    @Transactional
    public Map<String, Object> chat(Long userId, GuideChatRequest req) {
        GuideSession session;
        if (req.sessionId() == null) {
            session = new GuideSession();
            session.setUserId(userId);
            session.setTitle(req.question().length() > 60 ? req.question().substring(0, 60) : req.question());
            sessionMapper.insert(session);
        } else {
            session = sessionMapper.findById(req.sessionId(), userId);
            if (session == null) throw ApiException.notFound("Session not found");
        }

        GuideSession.GuideMessage userMsg = new GuideSession.GuideMessage();
        userMsg.setSessionId(session.getId());
        userMsg.setRole("user");
        userMsg.setContent(req.question());
        messageMapper.insert(userMsg);

        String systemPrompt = buildSystemPrompt(req.question());

        List<Map<String, String>> messages = new ArrayList<>();
        messages.add(Map.of("role", "system", "content", systemPrompt));
        for (GuideSession.GuideMessage m : messageMapper.findBySessionId(session.getId())) {
            messages.add(Map.of("role", m.getRole(), "content", m.getContent()));
        }

        String reply = qwen.chat("qwen-turbo", messages);

        GuideSession.GuideMessage botMsg = new GuideSession.GuideMessage();
        botMsg.setSessionId(session.getId());
        botMsg.setRole("assistant");
        botMsg.setContent(reply);
        messageMapper.insert(botMsg);

        return Map.of("sessionId", session.getId(), "reply", reply);
    }

    private String buildSystemPrompt(String question) {
        StringBuilder sb = new StringBuilder("""
                You are a knowledgeable AI tour guide for Guangdong / Lingnan culture. \
                Answer the user's question about travel, history, or culture. \
                Keep answers concise (under 100 words) and engaging.""");
        String keyword = question.length() > 20 ? question.substring(0, 20) : question;
        List<Knowledge> hits = knowledgeMapper.searchByKeyword(keyword, 3);
        if (!hits.isEmpty()) {
            sb.append(" Relevant cultural background:");
            for (Knowledge k : hits) {
                sb.append(" [").append(k.getTitleEn()).append(": ").append(k.getContentEn()).append("]");
            }
        }
        return sb.toString();
    }

    public GuideSession getSession(Long userId, Long id) {
        GuideSession session = sessionMapper.findById(id, userId);
        if (session == null) throw ApiException.notFound("Session not found");
        session.setMessages(messageMapper.findBySessionId(id));
        return session;
    }
}
