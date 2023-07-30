package parser

import (
	"bytes"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func TestParser_ParseAndEnhanceMetricsInto(t *testing.T) {
	type fields struct {
		kCli                  *kubernetes.Clientset
		nodeName              string
		podsMap               map[string]corev1.Pod
		enrichTags            map[string]string
		enrichFromLabels      bool
		enrichFromAnnotations bool
		filterPrefixes        []string
		replacePrefixes       map[string]string

		shouldValidateMetricLabel bool
	}

	type args struct {
		b []byte
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantWriter string
		wantErr    bool
	}{
		{
			name: "test 1: without any enrichment",
			fields: fields{
				kCli:     nil,
				nodeName: "test",
				podsMap: map[string]corev1.Pod{
					"test/test": {
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"without-group-1": "without-group-1-value",
								"group.io/test_1": "test_value_1",
								"group.io/test_2": "test_value_2",
							},
						},
					},
				},
				enrichFromLabels:      false,
				enrichFromAnnotations: false,
				filterPrefixes:        []string{},
				replacePrefixes:       map[string]string{},
			},
			args: args{
				b: []byte(`#HELP pod_cpu_usage_seconds_total Total user and system CPU time spent in seconds.
pod_memory_working_set_bytes{namespace="test",pod="test"} 827392 1689181712563`),
			},
			wantWriter: `#HELP pod_cpu_usage_seconds_total Total user and system CPU time spent in seconds.
pod_memory_working_set_bytes{namespace="test",pod="test"} 827392 1689181712563`,
			// wantWriter: `pod_memory_working_set_bytes{namespace="test",pod="test"} 827392 1689181712563`,
			wantErr: false,
		},

		{
			name: "test 2: enriching with labels",
			fields: fields{
				kCli:     nil,
				nodeName: "test",
				podsMap: map[string]corev1.Pod{
					"test/test": {
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"k1": "v1",
							},
						},
					},
				},
				enrichFromLabels:      true,
				enrichFromAnnotations: false,
				filterPrefixes:        []string{},
				replacePrefixes:       map[string]string{},
			},
			args: args{
				b: []byte(`pod_memory_working_set_bytes{namespace="test",pod="test",container="main"} 827392 1689181712563`),
			},
			wantWriter: `pod_memory_working_set_bytes{container="main",namespace="test",pod="test",k1="v1"} 827392 1689181712563`,
			wantErr:    false,
		},

		{
			name: "test 3: enriching with annotations",
			fields: fields{
				kCli:     nil,
				nodeName: "test",
				podsMap: map[string]corev1.Pod{
					"test/test": {
						ObjectMeta: metav1.ObjectMeta{
							Annotations: map[string]string{
								"k1": "v1",
							},
						},
					},
				},
				enrichFromLabels:      false,
				enrichFromAnnotations: true,
				filterPrefixes:        []string{},
				replacePrefixes:       map[string]string{},
			},
			args: args{
				b: []byte(`pod_memory_working_set_bytes{namespace="test",pod="test"} 827392 1689181712563`),
			},
			wantWriter: `pod_memory_working_set_bytes{namespace="test",pod="test",k1="v1"} 827392 1689181712563`,
			wantErr:    false,
		},

		{
			name: "test 4: enriching with custom tags",
			fields: fields{
				kCli:     nil,
				nodeName: "test",
				podsMap: map[string]corev1.Pod{
					"test/test": {
						ObjectMeta: metav1.ObjectMeta{
							Annotations: map[string]string{
								"k1": "v1",
							},
						},
					},
				},
				enrichTags:            map[string]string{"k2": "v2"},
				enrichFromLabels:      false,
				enrichFromAnnotations: false,
				filterPrefixes:        []string{},
				replacePrefixes:       map[string]string{},
			},
			args: args{
				b: []byte(`pod_memory_working_set_bytes{namespace="test",pod="test"} 827392 1689181712563`),
			},
			wantWriter: `pod_memory_working_set_bytes{namespace="test",pod="test",k2="v2"} 827392 1689181712563`,
			wantErr:    false,
		},

		{
			name: "test 5: enriching with labels,annotations and custom tags",
			fields: fields{
				kCli:     nil,
				nodeName: "test",
				podsMap: map[string]corev1.Pod{
					"test/test": {
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"label1": "label-value-1",
							},
							Annotations: map[string]string{
								"ann1": "ann-value-1",
							},
						},
					},
				},
				enrichTags:            map[string]string{"k1": "v1"},
				enrichFromLabels:      true,
				enrichFromAnnotations: true,
				filterPrefixes:        []string{},
				replacePrefixes:       map[string]string{},
			},
			args: args{
				b: []byte(`pod_memory_working_set_bytes{namespace="test",pod="test"} 827392 1689181712563`),
			},
			wantWriter: `pod_memory_working_set_bytes{namespace="test",pod="test",label1="label-value-1",ann1="ann-value-1",k1="v1"} 827392 1689181712563`,
			wantErr:    false,
		},

		{
			name: "test 6: enriching with labels, with filtering prefixes",
			fields: fields{
				kCli:     nil,
				nodeName: "test",
				podsMap: map[string]corev1.Pod{
					"test/test": {
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"label1":           "label-value-1",
								"group.io/label_1": "label-group-value-1",
							},
						},
					},
				},
				enrichFromLabels:      true,
				enrichFromAnnotations: false,
				filterPrefixes:        []string{"group.io/"},
				replacePrefixes:       map[string]string{},
			},
			args: args{
				b: []byte(`pod_memory_working_set_bytes{namespace="test",pod="test"} 827392 1689181712563`),
			},
			wantWriter: `pod_memory_working_set_bytes{namespace="test",pod="test",group.io/label_1="label-group-value-1"} 827392 1689181712563`,
			wantErr:    false,
		},

		{
			name: "test 7: enriching with annotations, with filtering prefixes",
			fields: fields{
				kCli:     nil,
				nodeName: "test",
				podsMap: map[string]corev1.Pod{
					"test/test": {
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"label1":           "label-value-1",
								"group.io/label_1": "label-group-value-1",
							},
							Annotations: map[string]string{
								"ann1":           "ann-value-1",
								"group.io/ann_1": "ann-group-value-1",
							},
						},
					},
				},
				enrichFromLabels:      false,
				enrichFromAnnotations: true,
				filterPrefixes:        []string{"group.io/"},
				replacePrefixes:       map[string]string{},
			},
			args: args{
				b: []byte(`pod_memory_working_set_bytes{namespace="test",pod="test"} 827392 1689181712563`),
			},
			wantWriter: `pod_memory_working_set_bytes{namespace="test",pod="test",group.io/ann_1="ann-group-value-1"} 827392 1689181712563`,
			wantErr:    false,
		},

		{
			name: "test 8: enriching with custom tags, with filtering prefixes",
			fields: fields{
				kCli:     nil,
				nodeName: "test",
				podsMap: map[string]corev1.Pod{
					"test/test": {},
				},
				enrichFromLabels:      false,
				enrichFromAnnotations: false,
				enrichTags:            map[string]string{"k1": "v1"},
				filterPrefixes:        []string{"group.io/"},
				replacePrefixes:       map[string]string{},
			},
			args: args{
				b: []byte(`pod_memory_working_set_bytes{namespace="test",pod="test"} 827392 1689181712563`),
			},
			wantWriter: `pod_memory_working_set_bytes{namespace="test",pod="test",k1="v1"} 827392 1689181712563`,
			wantErr:    false,
		},

		{
			name: "test 9: enriching with labels, annotations and custom tags, with filtering prefixes",
			fields: fields{
				kCli:     nil,
				nodeName: "test",
				podsMap: map[string]corev1.Pod{
					"test/test": {
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"label1":           "label-value-1",
								"group.io/label_1": "label-group-value-1",
							},
							Annotations: map[string]string{
								"ann1":           "ann-value-1",
								"group.io/ann_1": "ann-group-value-1",
							},
						},
					},
				},
				enrichFromLabels:      true,
				enrichFromAnnotations: true,
				enrichTags:            map[string]string{"k1": "v1"},
				filterPrefixes:        []string{"group.io/"},
				replacePrefixes:       map[string]string{},
			},
			args: args{
				b: []byte(`pod_memory_working_set_bytes{namespace="test",pod="test"} 827392 1689181712563`),
			},
			wantWriter: `pod_memory_working_set_bytes{namespace="test",pod="test",group.io/label_1="label-group-value-1",group.io/ann_1="ann-group-value-1",k1="v1"} 827392 1689181712563`,
			wantErr:    false,
		},

		{
			name: "test 10: enriching with labels, with filtering prefixes, and with replacing prefixes",
			fields: fields{
				kCli:     nil,
				nodeName: "test",
				podsMap: map[string]corev1.Pod{
					"test/test": {
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"label1":           "label-value-1",
								"group.io/label_1": "label-group-value-1",
							},
							Annotations: map[string]string{
								"ann1":           "ann-value-1",
								"group.io/ann_1": "ann-group-value-1",
							},
						},
					},
				},
				enrichFromLabels:      true,
				enrichFromAnnotations: false,
				filterPrefixes:        []string{"group.io/"},
				replacePrefixes:       map[string]string{"group.io/": "grp_"},
			},
			args: args{
				b: []byte(`pod_memory_working_set_bytes{namespace="test",pod="test"} 827392 1689181712563`),
			},
			wantWriter: `pod_memory_working_set_bytes{namespace="test",pod="test",grp_label_1="label-group-value-1"} 827392 1689181712563`,
			wantErr:    false,
		},

		{
			name: "test 11: enriching with annotations, filtering prefixes, and with replacing prefixes",
			fields: fields{
				kCli:     nil,
				nodeName: "test",
				podsMap: map[string]corev1.Pod{
					"test/test": {
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"label1":           "label-value-1",
								"group.io/label_1": "label-group-value-1",
							},
							Annotations: map[string]string{
								"ann1":           "ann-value-1",
								"group.io/ann_1": "ann-group-value-1",
							},
						},
					},
				},
				enrichFromLabels:      false,
				enrichFromAnnotations: true,
				filterPrefixes:        []string{"group.io/"},
				replacePrefixes:       map[string]string{"group.io/": "grp_"},
			},
			args: args{
				b: []byte(`pod_memory_working_set_bytes{namespace="test",pod="test"} 827392 1689181712563`),
			},
			wantWriter: `pod_memory_working_set_bytes{namespace="test",pod="test",grp_ann_1="ann-group-value-1"} 827392 1689181712563`,
			wantErr:    false,
		},

		{
			name: "test 12: enriching with custom tags, filtering prefixes, and with replacing prefixes",
			fields: fields{
				kCli:     nil,
				nodeName: "test",
				podsMap: map[string]corev1.Pod{
					"test/test": {},
				},
				enrichFromLabels:      false,
				enrichFromAnnotations: false,
				enrichTags:            map[string]string{"k1": "v1"},
				filterPrefixes:        []string{"group.io/"},
				replacePrefixes:       map[string]string{"group.io/": "grp_"},
			},
			args: args{
				b: []byte(`pod_memory_working_set_bytes{namespace="test",pod="test"} 827392 1689181712563`),
			},
			wantWriter: `pod_memory_working_set_bytes{namespace="test",pod="test",k1="v1"} 827392 1689181712563`,
			wantErr:    false,
		},

		{
			name: "test 13: enriching from labels, annotations and custom tags, with filtering prefixes and replacing prefixes",
			fields: fields{
				kCli:     nil,
				nodeName: "test",
				podsMap: map[string]corev1.Pod{
					"test/test": {
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"label1":           "label-value-1",
								"group.io/label_1": "label-group-value-1",
							},
							Annotations: map[string]string{
								"ann1":           "ann-value-1",
								"group.io/ann_1": "ann-group-value-1",
							},
						},
					},
				},
				enrichFromLabels:      true,
				enrichFromAnnotations: true,
				enrichTags:            map[string]string{"k1": "v1"},
				filterPrefixes:        []string{"group.io/"},
				replacePrefixes:       map[string]string{"group.io/": "grp_"},
			},
			args: args{
				b: []byte(`pod_memory_working_set_bytes{namespace="test",pod="test"} 827392 1689181712563`),
			},
			wantWriter: `pod_memory_working_set_bytes{namespace="test",pod="test",grp_label_1="label-group-value-1",grp_ann_1="ann-group-value-1",k1="v1"} 827392 1689181712563`,
			wantErr:    false,
		},

		{
			name: "test 14: with malformed metrics",
			fields: fields{
				kCli:     nil,
				nodeName: "test",
				podsMap: map[string]corev1.Pod{
					"test/test": {
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"label1":           "label-value-1",
								"group.io/label_1": "label-group-value-1",
							},
							Annotations: map[string]string{
								"ann1":           "ann-value-1",
								"group.io/ann_1": "ann-group-value-1",
							},
						},
					},
				},
				enrichFromLabels:      true,
				enrichFromAnnotations: true,
				filterPrefixes:        []string{"group.io/"},
				replacePrefixes:       map[string]string{"group.io/": "grp_"},
			},
			args: args{
				b: []byte(`pod_memory_working_set_bytes}namespace="test",pod="test"{ 827392 1689181712563`),
			},
			wantWriter: `pod_memory_working_set_bytes}namespace="test",pod="test"{ 827392 1689181712563`,
			wantErr:    false,
		},

		{
			name: "test 15: with malformed metrics, without any '{' or '}'",
			fields: fields{
				kCli:     nil,
				nodeName: "test",
				podsMap: map[string]corev1.Pod{
					"test/test": {
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"label1":           "label-value-1",
								"group.io/label_1": "label-group-value-1",
							},
							Annotations: map[string]string{
								"ann1":           "ann-value-1",
								"group.io/ann_1": "ann-group-value-1",
							},
						},
					},
				},
				enrichFromLabels:      true,
				enrichFromAnnotations: true,
				filterPrefixes:        []string{"group.io/"},
				replacePrefixes:       map[string]string{"group.io/": "grp_"},
			},
			args: args{
				b: []byte(`scrape error 0`),
			},
			wantWriter: `scrape error 0`,
			wantErr:    false,
		},

		{
			name: "test 16: with malformed metrics, without any closing '{'",
			fields: fields{
				kCli:     nil,
				nodeName: "test",
				podsMap: map[string]corev1.Pod{
					"test/test": {
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"label1":           "label-value-1",
								"group.io/label_1": "label-group-value-1",
							},
							Annotations: map[string]string{
								"ann1":           "ann-value-1",
								"group.io/ann_1": "ann-group-value-1",
							},
						},
					},
				},
				enrichFromLabels:      true,
				enrichFromAnnotations: true,
				filterPrefixes:        []string{"group.io/"},
				replacePrefixes:       map[string]string{"group.io/": "grp_"},
			},
			args: args{
				b: []byte(`pod_memory_working_set_bytes{namespace="test",pod="test" 827392 1689181712563`),
			},
			wantWriter: `pod_memory_working_set_bytes{namespace="test",pod="test" 827392 1689181712563`,
			wantErr:    false,
		},

		{
			name: "test 17: with malformed metrics, without any opening '}'",
			fields: fields{
				kCli:     nil,
				nodeName: "test",
				podsMap: map[string]corev1.Pod{
					"test/test": {
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"label1":           "label-value-1",
								"group.io/label_1": "label-group-value-1",
							},
							Annotations: map[string]string{
								"ann1":           "ann-value-1",
								"group.io/ann_1": "ann-group-value-1",
							},
						},
					},
				},
				enrichFromLabels:      true,
				enrichFromAnnotations: true,
				filterPrefixes:        []string{"group.io/"},
				replacePrefixes:       map[string]string{"group.io/": "grp_"},
			},
			args: args{
				b: []byte(`pod_memory_working_set_bytesnamespace="test",pod="test"} 827392 1689181712563`),
			},
			wantWriter: `pod_memory_working_set_bytesnamespace="test",pod="test"} 827392 1689181712563`,
			wantErr:    false,
		},

		{
			name: "test 18: with metric label regex validation enabled",
			fields: fields{
				kCli:     nil,
				nodeName: "test",
				podsMap: map[string]corev1.Pod{
					"test/test": {
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"label1":           "label-value-1",
								"group.io/label_1": "label-group-value-1",
							},
							Annotations: map[string]string{
								"ann1":             "ann-value-1",
								"group.io/ann_1":   "ann-group-value-1",
								"group.io/ann-2":   "ann-group-value-2",
								"group.io/ann_key": "ann-group-value-3",
							},
						},
					},
				},
				enrichFromLabels:          true,
				enrichFromAnnotations:     true,
				filterPrefixes:            []string{"group.io/"},
				replacePrefixes:           map[string]string{"group.io/": "grp_"},
				shouldValidateMetricLabel: true,
			},
			args: args{
				b: []byte(`pod_memory_working_set_bytes{namespace="test",pod="test"} 827392 1689181712563`),
			},
			wantWriter: `pod_memory_working_set_bytes{namespace="test",pod="test",grp_label_1="label-group-value-1",grp_ann_1="ann-group-value-1",grp_ann_key="ann-group-value-3"} 827392 1689181712563`,
			wantErr:    false,
		},
		{
			name: "test 19: given go template like value in enrich-tags",
			fields: fields{
				kCli:     nil,
				nodeName: "test",
				podsMap: map[string]corev1.Pod{
					"test/test": {
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"label1":           "label-value-1",
								"group.io/label_1": "label-group-value-1",
							},
							Annotations: map[string]string{
								"ann1": "ann-value-1",
							},
						},
					},
				},
				enrichFromLabels:      false,
				enrichFromAnnotations: false,
				enrichTags: map[string]string{
					"example_from_label": "{{ index .Labels \"label1\" }}",
					"example_from_ann":   "{{ index .Annotations \"ann1\" }}",
				},
				filterPrefixes:            []string{"group.io/"},
				replacePrefixes:           map[string]string{"group.io/": "grp_"},
				shouldValidateMetricLabel: true,
			},
			args: args{
				b: []byte(`pod_memory_working_set_bytes{namespace="test",pod="test"} 827392 1689181712563`),
			},
			wantWriter: `pod_memory_working_set_bytes{namespace="test",pod="test",example_from_label="label-value-1",example_from_ann="ann-value-1"} 827392 1689181712563`,
			wantErr:    false,
		},
	}
	for _, _tt := range tests {
		tt := _tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p, err := NewParser(tt.fields.kCli, tt.fields.nodeName, ParserOpts{
				PodsMap:                   tt.fields.podsMap,
				EnrichTags:                tt.fields.enrichTags,
				EnrichFromLabels:          tt.fields.enrichFromLabels,
				EnrichFromAnnotations:     tt.fields.enrichFromAnnotations,
				FilterPrefixes:            tt.fields.filterPrefixes,
				ReplacePrefixes:           tt.fields.replacePrefixes,
				ShouldValidateMetricLabel: tt.fields.shouldValidateMetricLabel,
				ValidLabelRegexExpr:       `^[a-zA-Z_][a-zA-Z0-9_]*$`, // source: https://prometheus.io/docs/concepts/data_model/
			})
			if err != nil {
				t.Error(err)
				return
			}

			writer := &bytes.Buffer{}
			if err := p.ParseAndEnhanceMetricsInto(tt.args.b, writer); (err != nil) != tt.wantErr {
				t.Errorf("Parser.ParseAndEnhanceMetricsInto() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotWriter := writer.String(); strings.TrimSpace(gotWriter) != strings.TrimSpace(tt.wantWriter) {
				t.Errorf("Parser.ParseAndEnhanceMetricsInto() = %v, want %v", gotWriter, tt.wantWriter)
			}
		})
	}
}
