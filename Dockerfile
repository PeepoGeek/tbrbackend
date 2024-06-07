# Usa una imagen base oficial de Go
FROM golang:1.22-alpine

# Establece el directorio de trabajo dentro del contenedor
WORKDIR /app

# Copia el archivo go.mod y go.sum
COPY go.mod go.sum ./

# Descarga los módulos Go
RUN go mod download

# Copia el código fuente de la aplicación
COPY . .

# Compila la aplicación
RUN go build -o main .

# Expone el puerto en el que la aplicación escuchará
EXPOSE 8080

# Comando para ejecutar la aplicación
CMD ["./main"]
