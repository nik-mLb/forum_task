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

CREATE TABLE forum.forum_user (
    forum TEXT NOT NULL REFERENCES forum.forum(slug),
    nickname TEXT NOT NULL REFERENCES forum."user"(nickname),
    PRIMARY KEY (forum, nickname)
);

--Индексы

-- Индексы для таблицы forum.user
CREATE INDEX user_lower_nickname_idx ON forum."user" (LOWER(nickname));
CREATE INDEX user_lower_email_idx ON forum."user" (LOWER(email));

-- Индексы для таблицы forum.forum
CREATE INDEX forum_lower_slug_idx ON forum.forum (LOWER(slug));

-- Индексы для таблицы forum.thread
CREATE INDEX thread_forum_idx ON forum.thread (forum);
CREATE INDEX thread_created_idx ON forum.thread (created);
CREATE INDEX thread_lower_slug_idx ON forum.thread (LOWER(slug));
CREATE INDEX thread_forum_created_idx ON forum.thread (forum, created);
CREATE INDEX thread_lower_forum_created_idx ON forum.thread (LOWER(forum), created);

-- Индексы для таблицы forum.post
CREATE INDEX post_thread_idx ON forum.post (thread);
CREATE INDEX post_created_idx ON forum.post (created);
CREATE INDEX post_path_idx ON forum.post USING GIN (path);
CREATE INDEX post_parent_idx ON forum.post (parent);
CREATE INDEX post_thread_id_idx ON forum.post (thread, id);
CREATE INDEX post_path1_idx ON forum.post ((path[1]));
CREATE INDEX post_thread_path1_idx ON forum.post (thread, (path[1]));
CREATE INDEX post_thread_parent_id_idx ON forum.post (thread, parent, id);

-- Индексы для таблицы forum.vote
CREATE INDEX vote_thread_idx ON forum.vote (thread);

-- Индексы для таблицы forum.forum_user
CREATE INDEX forum_user_lower_forum_nickname_idx ON forum.forum_user (LOWER(forum), nickname);
CREATE INDEX forum_user_nickname_idx ON forum.forum_user (nickname);

-- Триггеры

-- Обновление голосов
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

-- Обновление счетчика тредов в форуме
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

-- Установка пути для поста
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

-- Триггеры для поддержки forum_user

-- При вставке в thread
CREATE OR REPLACE FUNCTION add_forum_user_thread() RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO forum.forum_user(forum, nickname)
    VALUES (NEW.forum, NEW.author)
    ON CONFLICT (forum, nickname) DO NOTHING;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_add_forum_user_thread
AFTER INSERT ON forum.thread
FOR EACH ROW EXECUTE PROCEDURE add_forum_user_thread();