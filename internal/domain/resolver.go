package domain

import "sort"

func ResolveRelationships(raw RawExport) []Post {
	tagMap := make(map[string]Tag, len(raw.Tags))
	for _, t := range raw.Tags {
		tagMap[t.ID] = t
	}

	authorMap := make(map[string]Author, len(raw.Authors))
	for _, a := range raw.Authors {
		authorMap[a.ID] = a
	}

	metaMap := make(map[string]PostMeta, len(raw.PostsMeta))
	for _, m := range raw.PostsMeta {
		metaMap[m.PostID] = m
	}

	postTags := groupAndSort(raw.PostsTags, func(pt PostTag) (string, string, int) {
		return pt.PostID, pt.TagID, pt.SortOrder
	})
	postAuthors := groupAndSort(raw.PostsAuthors, func(pa PostAuthor) (string, string, int) {
		return pa.PostID, pa.AuthorID, pa.SortOrder
	})

	result := make([]Post, 0, len(raw.Posts))
	for _, rp := range raw.Posts {
		p := rp.Post

		if tagIDs, ok := postTags[p.ID]; ok {
			p.Tags = make([]Tag, 0, len(tagIDs))
			for _, id := range tagIDs {
				if tag, found := tagMap[id]; found {
					p.Tags = append(p.Tags, tag)
				}
			}
		}

		if authorIDs, ok := postAuthors[p.ID]; ok {
			p.Authors = make([]Author, 0, len(authorIDs))
			for _, id := range authorIDs {
				if author, found := authorMap[id]; found {
					p.Authors = append(p.Authors, author)
				}
			}
		}

		if meta, ok := metaMap[p.ID]; ok {
			if meta.MetaDescription != "" {
				p.MetaDescription = meta.MetaDescription
			}
			if meta.MetaTitle != "" {
				p.MetaTitle = meta.MetaTitle
			}
			if meta.FeatureImageAlt != "" {
				p.FeatureImageAlt = meta.FeatureImageAlt
			}
			if meta.FeatureImageCaption != "" {
				p.FeatureImageCaption = meta.FeatureImageCaption
			}
		}

		result = append(result, p)
	}

	return result
}

type sortedRelation struct {
	ID        string
	SortOrder int
}

func groupAndSort[T any](
	rows []T,
	extractFn func(T) (postID string, relatedID string, sortOrder int),
) map[string][]string {
	grouped := make(map[string][]sortedRelation)
	for _, row := range rows {
		pid, rid, order := extractFn(row)
		grouped[pid] = append(grouped[pid], sortedRelation{ID: rid, SortOrder: order})
	}

	result := make(map[string][]string, len(grouped))
	for pid, rels := range grouped {
		sort.Slice(rels, func(i, j int) bool {
			return rels[i].SortOrder < rels[j].SortOrder
		})
		ids := make([]string, len(rels))
		for i, r := range rels {
			ids[i] = r.ID
		}
		result[pid] = ids
	}
	return result
}
