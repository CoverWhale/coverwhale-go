// Copyright 2023 Cover Whale Insurance Solutions Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tpl

func EdgeDBInfra() []byte {
	return []byte(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: edgedb
  labels:
    app: edgedb
spec:
  replicas: 1
  selector:
    matchLabels:
      app: edgedb
  template:
    metadata:
      labels:
        app: edgedb
    spec:
      enableServiceLinks: false
      containers:
        - name: edgedb
          image: edgedb/edgedb:3
          ports:
            - containerPort: 5656
          readinessProbe:
            httpGet:
              path: /server/status/ready
              port: 5656
          livenessProbe:
            httpGet:
              path: /server/status/ready
              port: 5656
          env:
            - name: EDGEDB_SERVER_SECURITY
              value: insecure_dev_mode
            - name: EDGEDB_SERVER_ADMIN_UI
              value: enabled
            - name: EDGEDB_SERVER_BACKEND_DSN
              value: "postgresql://edgedb:edgedb@postgres:5432"
          volumeMounts:
            - name: schema
              mountPath: /dbschema
      volumes:
        - name: schema
          hostPath:
            path: /dbschema
---
apiVersion: v1
kind: Service
metadata:
  name: edgedb
spec:
  selector:
    app: edgedb
  type: ClusterIP
  ports:
    - protocol: TCP
      port: 5656
      targetPort: 5656
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: edgedb
  labels:
    name: edgeb
  annotations:
    kubernetes.io/ingress.class: traefik
spec:
  rules:
  - host: edgedb.127.0.0.1.nip.io
    http:
      paths:
      - pathType: Prefix
        path: "/"
        backend:
          service:
            name: edgedb
            port: 
              number: 5656
---
apiVersion: v1
kind: PersistentVolume
metadata:
  name: postgres
  labels:
    type: local
spec:
  capacity:
    storage: 1Gi
  volumeMode: Filesystem
  accessModes:
  - ReadWriteOnce
  persistentVolumeReclaimPolicy: Retain
  storageClassName: "local-path"
  hostPath:
    path: /data
---
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: postgres
spec:
  storageClassName: "local-path"
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: postgres
spec:
  selector:
    matchLabels:
      app: postgres
  serviceName: postgres
  replicas: 1
  template:
    metadata:
      labels:
        app: postgres
    spec:
      containers:
      - name: postgres
        image: postgres:latest
        envFrom:
        - configMapRef:
            name: postgres
        ports:
        - containerPort: 5432
          name: postgres
        volumeMounts:
        - name: data
          mountPath: /var/lib/postgresql/data
      volumes:
      - name: data
        persistentVolumeClaim:
          claimName: postgres
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: postgres
data:
  POSTGRES_DB: edgedb
  POSTGRES_USER: edgedb
  POSTGRES_PASSWORD: edgedb
---
apiVersion: v1
kind: Service
metadata:
  name: postgres
spec:
  selector:
    app: postgres
  type: ClusterIP
  ports:
  - port: 5432
    targetPort: 5432


`)
}

func NATSInfra() []byte {
	return []byte(`
config:
  cluster:
    enabled: true
  jetstream:
    enabled: true
    fileStore:
      enabled: false
      pvc:
        enabled: false
      maxSize: 1Gi
  natsBox:
    enabled: true
`)
}
