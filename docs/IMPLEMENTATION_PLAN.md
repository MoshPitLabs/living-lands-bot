# Living Lands Discord Bot - Implementation Plan

## Overview

This document outlines the phased implementation approach for the Living Lands Discord Bot. Each phase builds upon the previous, delivering incremental value.

**Estimated Total Duration:** 4-6 weeks (part-time development)
**Current Status:** Phase 5 Complete - MVP Features Implemented

---

## Phase 0: Foundation (Week 1)

**Goal:** Set up project structure and core infrastructure

### Tasks

#### 0.1 Project Setup
- [ ] Initialize Go module
- [ ] Set up project directory structure
- [ ] Create .gitignore and .env.example
- [ ] Set up Docker Compose configuration
- [ ] Create basic Dockerfile
- [ ] Set up CI/CD pipeline (GitHub Actions)
- [ ] Configure linting (golangci-lint)
- [ ] Set up pre-commit hooks

#### 0.2 Core Dependencies
- [ ] Add discordgo v0.29.0
- [ ] Add Fiber v2.52+
- [ ] Add GORM v1.25+
- [ ] Add go-redis v9.4+
- [ ] Add chroma-go v0.1+
- [ ] Add go-playground/validator v10+
- [ ] Add envconfig v1.4+

#### 0.3 Configuration & Logging
- [ ] Implement environment configuration (envconfig)
- [ ] Set up slog with JSON handler
- [ ] Create configuration validation
- [ ] Add graceful shutdown handling
- [ ] Implement health check endpoint

### Deliverables
- Working Docker Compose setup
- Bot connects to Discord gateway
- Health endpoint responds

### Success Criteria
- [ ] `docker compose up -d` starts all services
- [ ] Bot logs successful connection to Discord
- [ ] `curl http://localhost:8000/health` returns 200

---

## Phase 1: Database & Models (Week 1-2)

**Goal:** Set up database schema and basic CRUD operations

### Tasks

#### 1.1 Database Setup
- [ ] Configure GORM connection to PostgreSQL
- [ ] Set up golang-migrate
- [ ] Create initial migration files
- [ ] Implement database health checks
- [ ] Add connection pooling configuration

#### 1.2 Database Models
- [ ] User model (Discord + Hytale linking)
- [ ] Conversation model (LLM context)
- [ ] ChannelRoute model (navigation)
- [ ] WelcomeTemplate model
- [ ] DocSource model (RAG metadata)

#### 1.3 Repository Layer
- [ ] UserRepository with CRUD operations
- [ ] ConversationRepository with TTL cleanup
- [ ] ChannelRouteRepository
- [ ] WelcomeTemplateRepository
- [ ] DocSourceRepository

#### 1.4 Database Seeding
- [ ] Create seed data for welcome templates
- [ ] Create seed data for channel routes
- [ ] Add migration runner script

### Deliverables
- Complete database schema
- Working migrations
- Repository tests

### Success Criteria
- [ ] All migrations run successfully
- [ ] Repository tests pass
- [ ] Seed data populates on first run

---

## Phase 2: Discord Bot Core (Week 2-3)

**Goal:** Implement basic Discord bot functionality

### Tasks

#### 2.1 Bot Initialization
- [ ] Create Bot struct with session management
- [ ] Implement gateway intent configuration
- [ ] Add event handler registration
- [ ] Set up command registration
- [ ] Implement reconnect logic

#### 2.2 Slash Commands
- [ ] `/link` - Generate verification code
- [ ] `/guide` - Show channel navigation
- [ ] `/ask` - Ask questions (placeholder for now)
- [ ] Command validation
- [ ] Error handling for commands

#### 2.3 Event Handlers
- [ ] GuildMemberAdd (welcome message)
- [ ] InteractionCreate (button/modal handling)
- [ ] MessageCreate (optional - for non-slash interactions)

#### 2.4 Welcome System
- [ ] Weighted random welcome message selection
- [ ] User mention formatting
- [ ] Welcome channel configuration
- [ ] Test welcome messages

#### 2.5 Channel Navigation
- [ ] Button-based channel guide
- [ ] Embed with channel descriptions
- [ ] Button interaction handling
- [ ] Dynamic channel mapping from database

### Deliverables
- Working slash commands
- Welcome messages for new members
- Interactive channel guide

### Success Criteria
- [ ] `/link` generates and stores verification code
- [ ] `/guide` shows interactive buttons
- [ ] New members receive welcome message
- [ ] Commands handle errors gracefully

---

## Phase 3: HTTP API & Account Linking (Week 3)

**Goal:** Implement Hytale webhook integration for account linking

### Tasks

#### 3.1 Fiber Server Setup
- [ ] Initialize Fiber app
- [ ] Configure middleware (logger, recover, CORS)
- [ ] Set up route groups
- [ ] Implement request validation

#### 3.2 Verification Endpoint
- [ ] POST /api/v1/verify endpoint
- [ ] API key authentication middleware
- [ ] Request validation (code, username, UUID)
- [ ] Link accounts in database
- [ ] Response formatting

#### 3.3 Account Linking Service
- [ ] GenerateVerificationCode method
- [ ] VerifyLink method
- [ ] Code expiration handling
- [ ] Duplicate prevention
- [ ] Audit logging

#### 3.4 Security
- [ ] HMAC/API key validation
- [ ] Rate limiting on verification endpoint
- [ ] Input sanitization
- [ ] Error response standardization

### Deliverables
- Working HTTP API
- End-to-end account linking
- Security implementation

### Success Criteria
- [ ] Hytale server can call webhook successfully
- [ ] Verification codes expire correctly
- [ ] Accounts link properly
- [ ] Unauthorized requests rejected

---

## Phase 4: LLM Integration (Week 3-4)

**Goal:** Integrate Ollama for intelligent responses

### Tasks

#### 4.1 Ollama Client
- [ ] HTTP client for Ollama API
- [ ] Generate endpoint wrapper
- [ ] Embeddings endpoint wrapper
- [ ] Error handling and retries
- [ ] Timeout configuration

#### 4.2 Personality System
- [ ] YAML personality configuration
- [ ] System prompt builder
- [ ] Lore-friendly response formatting
- [ ] Personality switching support (future-proofing)

#### 4.3 LLM Service
- [ ] GenerateResponse method
- [ ] Context window management
- [ ] Token limit handling
- [ ] Temperature and sampling configuration

#### 4.4 Rate Limiting
- [ ] Per-user rate limiting (Redis)
- [ ] Global rate limiting
- [ ] Rate limit error messages
- [ ] Rate limit bypass for admins

#### 4.5 Conversation Memory
- [ ] Redis-based conversation storage
- [ ] TTL-based cleanup
- [ ] Context retrieval
- [ ] Conversation thread management

### Deliverables
- Working LLM integration
- Rate limiting
- Conversation context

### Success Criteria
- [ ] Ollama responds to queries
- [ ] Rate limits enforced correctly
- [ ] Conversations maintain context
- [ ] Bot stays in character

---

## Phase 5: RAG Pipeline (Week 4-5)

**Goal:** Implement Retrieval-Augmented Generation for accurate mod answers

### Tasks

#### 5.1 Document Indexing
- [ ] Document parser (Markdown, TXT)
- [ ] Chunking strategy (semantic chunks)
- [ ] Indexing script
- [ ] Incremental updates
- [ ] Checksum-based change detection

#### 5.2 ChromaDB Integration
- [ ] Collection setup
- [ ] Embedding generation (nomic-embed-text)
- [ ] Document storage
- [ ] Metadata management
- [ ] Query interface

#### 5.3 RAG Service
- [ ] Query embedding generation
- [ ] Vector search implementation
- [ ] Context assembly
- [ ] Relevance scoring
- [ ] Fallback to general knowledge

#### 5.4 Document Management
- [ ] DocSource tracking
- [ ] Re-indexing workflow
- [ ] Document versioning
- [ ] Source attribution in responses

#### 5.5 `/ask` Command Completion
- [ ] Integrate RAG with LLM service
- [ ] Source citation in responses
- [ ] "I don't know" handling
- [ ] Confidence thresholding

### Deliverables
- Document indexing system
- RAG-powered Q&A
- Source attribution

### Success Criteria
- [ ] Documents index successfully
- [ ] RAG retrieves relevant context
- [ ] Answers cite sources
- [ ] Handles unknown questions gracefully

---

## Phase 6: Testing & Polish (Week 5-6)

**Goal:** Comprehensive testing and production readiness

### Tasks

#### 6.1 Unit Testing
- [ ] Service layer tests
- [ ] Repository tests
- [ ] Utility function tests
- [ ] Mock external dependencies
- [ ] Test coverage > 70%

#### 6.2 Integration Testing
- [ ] Discord bot integration tests
- [ ] API endpoint tests
- [ ] Database integration tests
- [ ] Ollama integration tests (mocked)

#### 6.3 Manual Testing
- [ ] Test all slash commands
- [ ] Test account linking flow
- [ ] Test rate limiting
- [ ] Test welcome messages
- [ ] Test channel navigation
- [ ] Test RAG Q&A

#### 6.4 Documentation
- [ ] API documentation
- [ ] Deployment guide
- [ ] Troubleshooting guide
- [ ] Environment variable reference
- [ ] Discord bot setup guide

#### 6.5 Production Readiness
- [ ] Add structured logging
- [ ] Implement metrics (optional)
- [ ] Add alerting webhooks (optional)
- [ ] Performance optimization
- [ ] Resource limit configuration

### Deliverables
- Test suite
- Complete documentation
- Production-ready deployment

### Success Criteria
- [ ] All tests pass
- [ ] Code coverage > 70%
- [ ] Documentation complete
- [ ] Bot runs stable for 24 hours
- [ ] No critical errors in logs

---

## Phase 7: Deployment & Launch (Week 6)

**Goal:** Deploy to production and go live

### Tasks

#### 7.1 Pre-Deployment
- [ ] Final code review
- [ ] Security audit
- [ ] Load testing (if applicable)
- [ ] Backup strategy verification
- [ ] Rollback plan

#### 7.2 Production Deployment
- [ ] Deploy to production server
- [ ] Configure production environment
- [ ] Set up monitoring
- [ ] Configure log aggregation
- [ ] SSL/TLS configuration (if needed)

#### 7.3 Launch
- [ ] Invite bot to Discord server
- [ ] Configure permissions
- [ ] Test in production
- [ ] Announce launch

#### 7.4 Post-Launch
- [ ] Monitor for issues
- [ ] Collect user feedback
- [ ] Fix critical bugs
- [ ] Document lessons learned

### Deliverables
- Production deployment
- Live bot in Discord server

### Success Criteria
- [ ] Bot online and responsive
- [ ] All features working in production
- [ ] No critical errors
- [ ] Users can interact successfully

---

## Implementation Schedule

| Phase | Duration | Start | End | Status |
|-------|----------|-------|-----|--------|
| Phase 0: Foundation | 1 week | Week 1 | Week 1 | Not Started |
| Phase 1: Database | 1 week | Week 1 | Week 2 | Not Started |
| Phase 2: Discord Core | 1 week | Week 2 | Week 3 | Not Started |
| Phase 3: API & Linking | 1 week | Week 3 | Week 3 | Not Started |
| Phase 4: LLM Integration | 1 week | Week 3 | Week 4 | Not Started |
| Phase 5: RAG Pipeline | 1 week | Week 4 | Week 5 | Not Started |
| Phase 6: Testing | 1 week | Week 5 | Week 6 | Not Started |
| Phase 7: Deployment | 3 days | Week 6 | Week 6 | Not Started |

---

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Ollama performance issues | Medium | High | Use smaller model, add caching |
| Discord rate limiting | Medium | Medium | Implement rate limiting, exponential backoff |
| Database connection issues | Low | High | Connection pooling, retry logic |
| Hytale API changes | Low | Medium | Versioned API, abstraction layer |
| ChromaDB compatibility | Low | Medium | Test thoroughly, have fallback |

---

## Definition of Done

A phase is considered complete when:

1. All tasks are implemented and tested
2. Code review completed
3. Documentation updated
4. Tests passing (unit + integration)
5. Deployed to staging and verified
6. No critical or high-priority bugs open

---

**Document Version:** 1.0  
**Last Updated:** 2026-01-31  
**Status:** Draft
