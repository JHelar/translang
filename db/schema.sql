create table if not exists "translation" (
    [id] integer not null primary key autoincrement,
    [figma_source_url] text not null unique,
    [context_image_url] text,
    [created_at] timestamp not null default current_timestamp,
    [synced_at] timestamp not null default current_timestamp
);
create table if not exists "translation_node" (
    [id] integer not null primary key autoincrement,
    [source_text] text not null,
    [copy_key] text not null unique,
    [created_at] timestamp not null default current_timestamp
);
create table if not exists "translation_to_translation_node" (
    [translation_id] integer not null references [translation]([id]) on delete cascade,
    [figma_text_node_id] text not null,
    [translation_node_id] integer not null references [translation_node]([id]) on delete cascade,
    primary key (translation_node_id, translation_id)
);
create table if not exists "translation_node_value" (
    [translation_node_id] integer not null references [translation_node]([id]) on delete cascade,
    [copy_language] text not null,
    [copy_text] text not null,
    [created_at] timestamp not null default current_timestamp,
    primary key (translation_node_id, copy_language)
);