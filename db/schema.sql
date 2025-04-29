create table if not exists "translation" (
    [id] integer not null primary key autoincrement,
    [figma_source_url] text not null,
    [context_image_url] text,
    [createdAt] timestamp not null default current_timestamp,
    [syncedAt] timestamp not null default current_timestamp
);

create table if not exists "translation_node" (
    [id] integer not null primary key autoincrement,
    [figma_text_node_id] text not null,
    [translation_id] integer not null references [translation]([id]) on delete cascade,
    [source_text] text not null,
);

create table if not exists "translation_node_value" (
    [translation_node_id] integer not null references [translation_node]([id]) on delete cascade,
    [copy_key] text not null,
    [copy_language] text not null,
    [copy_text] text not null,
    primary key (copy_key, copy_language)
);