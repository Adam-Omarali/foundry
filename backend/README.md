SQL TABLES:

user:
id UUID
email text

documents:
id UUID
user_id UUID FK (user.id)
url text
title text
summary text

embeddings:
id UUID
docmument_id UUID FK (documents.id)
vector VECTOR[]
