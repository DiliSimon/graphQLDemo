package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/graphql-go/graphql"
)

type Media struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"` //取值 Image 或 Video
}

var medias = []Media{
	{
		ID:   1,
		Name: "Canyon Image",
		Type: "Image",
	},
	{
		ID:   2,
		Name: "Sunset Video",
		Type: "Video",
	},
}

var mediaType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Media",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.Int,
			},
			"name": &graphql.Field{
				Type: graphql.String,
			},
			"type": &graphql.Field{
				Type: graphql.String,
			},
		},
	},
)

var queryType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"media": &graphql.Field{
				Type:        mediaType,
				Description: "Get media by id",
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.Int,
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					id, ok := p.Args["id"].(int)
					if ok {
						// Find product
						for _, m := range medias {
							if int(m.ID) == id {
								return m, nil
							}
						}
					}
					return nil, nil
				},
			},
			/* Get (read) media list
			   http://localhost:8080/product?query={mediaList{id,name,type}}
			*/
			"mediaList": &graphql.Field{
				Type:        graphql.NewList(mediaType),
				Description: "Get media list",
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					return medias, nil
				},
			},
		},
	})

var mutationType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Mutation",
	Fields: graphql.Fields{
		"create": &graphql.Field{
			Type:        mediaType,
			Description: "Create new media",
			Args: graphql.FieldConfigArgument{
				"name": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"type": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
			},
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				rand.Seed(time.Now().UnixNano())
				newMedia := Media{
					ID:   int64(rand.Intn(100000)),
					Name: params.Args["name"].(string),
					Type: params.Args["info"].(string),
				}
				medias = append(medias, newMedia)
				return newMedia, nil
			},
		},

		"update": &graphql.Field{
			Type:        mediaType,
			Description: "Update media by id",
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.Int),
				},
				"name": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"type": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
			},
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				id, _ := params.Args["id"].(int)
				name, nameOk := params.Args["name"].(string)
				mediaType, typeOK := params.Args["type"].(string)
				newMedia := Media{}
				for i, p := range medias {
					if int64(id) == p.ID {
						if nameOk {
							medias[i].Name = name
						}
						if typeOK {
							medias[i].Type = mediaType
						}
						newMedia = medias[i]
						break
					}
				}
				return newMedia, nil
			},
		},

		"delete": &graphql.Field{
			Type:        mediaType,
			Description: "Delete media by id",
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.Int),
				},
			},
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				id, _ := params.Args["id"].(int)
				media := Media{}
				for i, p := range medias {
					if int64(id) == p.ID {
						media = medias[i]
						medias = append(medias[:i], medias[i+1:]...)
					}
				}

				return media, nil
			},
		},
	},
})

var schema, _ = graphql.NewSchema(
	graphql.SchemaConfig{
		Query:    queryType,
		Mutation: mutationType,
	},
)

func executeQuery(query string, schema graphql.Schema) *graphql.Result {
	result := graphql.Do(graphql.Params{
		Schema:        schema,
		RequestString: query,
	})
	if len(result.Errors) > 0 {
		fmt.Printf("errors: %v", result.Errors)
	}
	return result
}

func main() {
	http.HandleFunc("/gqlMedia", func(w http.ResponseWriter, r *http.Request) {
		result := executeQuery(r.URL.Query().Get("query"), schema)
		json.NewEncoder(w).Encode(result)
	})

	fmt.Println("Server is running on port 8080")
	http.ListenAndServe(":8080", nil)
}
