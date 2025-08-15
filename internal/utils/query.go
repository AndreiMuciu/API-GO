package utils

import (
	"net/http"
	"strconv"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
)

// ListQuery holds Mongo-friendly query parts parsed from URL params.
type ListQuery struct {
    Filter bson.M
    Sort   bson.D
    Limit  int64
    Skip   int64
    Page   int64
}

// ParseListQuery parses common query params into a ListQuery.
// allowedFields maps allowed field names to a simple type: "string" or "int".
// allowedSort is a set (map[string]bool) of fields that can be sorted by.
func ParseListQuery(r *http.Request, allowedFields map[string]string, allowedSort map[string]bool, defaultSort string, defaultLimit, maxLimit int64) ListQuery {
    q := r.URL.Query()
    filter := bson.M{}

    // Build filters
    for field, typ := range allowedFields {
        // equality: field=value
        if val := strings.TrimSpace(q.Get(field)); val != "" {
            if typ == "int" {
                if iv, err := strconv.Atoi(val); err == nil {
                    filter[field] = iv
                }
            } else {
                filter[field] = val
            }
        }
        // contains (case-insensitive): field_like=value
        if val := strings.TrimSpace(q.Get(field + "_like")); val != "" {
            filter[field] = bson.M{"$regex": val, "$options": "i"}
        }
        // min/max for numeric
        if typ == "int" {
            if val := strings.TrimSpace(q.Get(field + "_min")); val != "" {
                if iv, err := strconv.Atoi(val); err == nil {
                    addRange(filter, field, "$gte", iv)
                }
            }
            if val := strings.TrimSpace(q.Get(field + "_max")); val != "" {
                if iv, err := strconv.Atoi(val); err == nil {
                    addRange(filter, field, "$lte", iv)
                }
            }
        }
    }

    // Global search across string fields: q=text
    if s := strings.TrimSpace(q.Get("q")); s != "" {
        ors := bson.A{}
        for f, t := range allowedFields {
            if t == "string" {
                ors = append(ors, bson.M{f: bson.M{"$regex": s, "$options": "i"}})
            }
        }
        if len(ors) > 0 {
            if len(filter) == 0 {
                filter = bson.M{"$or": ors}
            } else {
                filter = bson.M{"$and": bson.A{filter, bson.M{"$or": ors}}}
            }
        }
    }

    // Sorting: sort=field,-other
    sortSpec := bson.D{}
    if sortParam := q.Get("sort"); sortParam != "" {
        parts := strings.Split(sortParam, ",")
        for _, p := range parts {
            p = strings.TrimSpace(p)
            if p == "" { continue }
            dir := 1
            name := p
            if strings.HasPrefix(p, "-") {
                dir = -1
                name = strings.TrimPrefix(p, "-")
            }
            if allowedSort[name] {
                sortSpec = append(sortSpec, bson.E{Key: name, Value: dir})
            }
        }
    }
    if len(sortSpec) == 0 && defaultSort != "" {
        dir := 1
        name := defaultSort
        if strings.HasPrefix(defaultSort, "-") {
            dir = -1
            name = strings.TrimPrefix(defaultSort, "-")
        }
        if allowedSort[name] {
            sortSpec = append(sortSpec, bson.E{Key: name, Value: dir})
        }
    }

    // Pagination: page, limit
    page := int64(1)
    limit := defaultLimit
    if v := q.Get("page"); v != "" {
        if iv, err := strconv.Atoi(v); err == nil && iv > 0 {
            page = int64(iv)
        }
    }
    if v := q.Get("limit"); v != "" {
        if iv, err := strconv.Atoi(v); err == nil && iv > 0 {
            limit = int64(iv)
        }
    }
    if maxLimit > 0 && limit > maxLimit {
        limit = maxLimit
    }
    skip := (page - 1) * limit

    return ListQuery{Filter: filter, Sort: sortSpec, Limit: limit, Skip: skip, Page: page}
}

func addRange(m bson.M, field, op string, val int) {
    if existing, ok := m[field]; ok {
        if sub, ok2 := existing.(bson.M); ok2 {
            sub[op] = val
            m[field] = sub
            return
        }
    }
    m[field] = bson.M{op: val}
}
