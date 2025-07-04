CREATE SCHEMA forum;

CREATE TABLE forum."user" (
    id SERIAL PRIMARY KEY,
    nickname TEXT NOT NULL UNIQUE,
    fullname TEXT,
    about TEXT,
    email TEXT NOT NULL UNIQUE
);

CREATE TABLE forum.forum (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    "user" TEXT NOT NULL REFERENCES forum."user"(nickname),
    slug TEXT NOT NULL UNIQUE,
    posts INTEGER DEFAULT 0 NOT NULL,
    threads INTEGER DEFAULT 0 NOT NULL
);

CREATE TABLE forum.thread (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    author TEXT NOT NULL REFERENCES forum."user"(nickname),
    forum TEXT NOT NULL REFERENCES forum.forum(slug),
    message TEXT NOT NULL,
    votes INTEGER DEFAULT 0 NOT NULL,
    slug TEXT,
    created TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE forum.post (
    id SERIAL PRIMARY KEY,
    parent INTEGER DEFAULT 0 NOT NULL,
    author TEXT NOT NULL REFERENCES forum."user"(nickname),
    message TEXT NOT NULL,
    isEdited BOOLEAN DEFAULT FALSE NOT NULL,
    forum TEXT NOT NULL,
    thread INTEGER NOT NULL REFERENCES forum.thread(id),
    created TIMESTAMPTZ DEFAULT NOW(),
    path INTEGER[] NOT NULL
);

CREATE TABLE forum.vote (
    id SERIAL PRIMARY KEY,
    nickname TEXT NOT NULL REFERENCES forum."user"(nickname),
    voice SMALLINT NOT NULL,
    thread INTEGER NOT NULL REFERENCES forum.thread(id),
    UNIQUE (nickname, thread)
);

-- Индексы
CREATE INDEX idx_thread_forum ON forum.thread(forum);
CREATE INDEX idx_thread_created ON forum.thread(created);
CREATE INDEX idx_post_thread ON forum.post(thread);
CREATE INDEX idx_post_thread_parent ON forum.post(thread, parent);
CREATE INDEX idx_post_thread_id ON forum.post(thread, id);
CREATE INDEX idx_post_path ON forum.post USING GIN(path);
CREATE INDEX idx_post_thread_path ON forum.post(thread, path);
CREATE INDEX idx_post_path_root ON forum.post ((path[1])) WHERE parent = 0;
CREATE INDEX idx_post_thread_created ON forum.post(thread, created);
CREATE INDEX idx_post_parent_thread ON forum.post(parent, thread);

-- Триггеры
CREATE OR REPLACE FUNCTION update_thread_votes() RETURNS TRIGGER AS $$
BEGIN
    -- Если это UPDATE и голос изменился
    IF TG_OP = 'UPDATE' AND OLD.voice <> NEW.voice THEN
        UPDATE forum.thread 
        SET votes = votes - OLD.voice + NEW.voice
        WHERE id = NEW.thread;
    -- Если это INSERT
    ELSIF TG_OP = 'INSERT' THEN
        UPDATE forum.thread 
        SET votes = votes + NEW.voice
        WHERE id = NEW.thread;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_votes
AFTER INSERT OR UPDATE ON forum.vote
FOR EACH ROW EXECUTE PROCEDURE update_thread_votes();

CREATE OR REPLACE FUNCTION update_forum_threads() RETURNS TRIGGER AS $$
BEGIN
    UPDATE forum.forum 
    SET threads = threads + 1 
    WHERE slug = NEW.forum;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_threads
AFTER INSERT ON forum.thread
FOR EACH ROW EXECUTE PROCEDURE update_forum_threads();

CREATE OR REPLACE FUNCTION update_forum_posts() RETURNS TRIGGER AS $$
BEGIN
    UPDATE forum.forum 
    SET posts = posts + 1 
    WHERE slug = NEW.forum;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_posts
AFTER INSERT ON forum.post
FOR EACH ROW EXECUTE PROCEDURE update_forum_posts();

CREATE OR REPLACE FUNCTION set_post_path() RETURNS TRIGGER AS $$
DECLARE
    parent_path INTEGER[];
BEGIN
    IF NEW.parent = 0 THEN
        NEW.path := ARRAY[NEW.id];
    ELSE
        -- Проверяем, что родитель существует и в том же треде
        SELECT path INTO parent_path 
        FROM forum.post 
        WHERE id = NEW.parent AND thread = NEW.thread;
        
        IF NOT FOUND THEN
            RAISE EXCEPTION 'Parent post % not found in thread %', NEW.parent, NEW.thread;
        END IF;
        
        NEW.path := parent_path || NEW.id;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_set_post_path
BEFORE INSERT ON forum.post
FOR EACH ROW EXECUTE PROCEDURE set_post_path();