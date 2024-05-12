package main

import (
	"database/sql"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	// prepare
	// настраиваем подключение к tracker.db
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавляем новую посылку в БД, убеждаемся в отсутствии ошибки и наличии идентификатора
	id, err := store.Add(parcel)
	assert.NoError(t, err)
	require.NotZero(t, id)
	parcel.Number = id

	// get
	// получаем только что добавленную посылку, убеждаемся в отсутствии ошибки
	// проверяем, что значения всех полей в полученном объекте совпадают со значениями полей в переменной parcel
	batch, err := store.Get(parcel.Number)
	assert.NoError(t, err)
	assert.Equal(t, parcel, batch)

	// delete
	// удаляем добавленную посылку, убеждаемся в отсутствии ошибки
	err = store.Delete(parcel.Number)
	require.NoError(t, err)

	// проверяем, что посылку больше нельзя получить из БД
	_, err = store.Get(id)
	require.Error(t, err) //ожидаема ошибка так как запись с посылкой удалена

}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	// настраиваем подключение к tracker.db
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавляем новую посылку в БД, убеждаемся в отсутствии ошибки и наличии идентификатора
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id)

	parcel.Number = id
	// set address
	// обновляем адрес, убеждаемся в отсутствии ошибки
	newAddress := "new test address"
	err = store.SetAddress(parcel.Number, newAddress)
	require.NoError(t, err)

	// check
	// получаем добавленную посылку и убеждаемся, что адрес обновился
	batch, err := store.Get(id)
	assert.NoError(t, err)
	assert.Equal(t, newAddress, batch.Address)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	// настраиваем подключение к tracker.db
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавляем новую посылку в БД, убеждаемся в отсутствии ошибки и наличии идентификатора
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id)

	// set status
	// обновляем статус, убеждаемся в отсутствии ошибки
	newStatus := ParcelStatusSent
	err = store.SetStatus(id, newStatus)
	require.NoError(t, err)

	// check
	// получаем добавленную посылку и убеждаемся, что статус обновился
	updBatch, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, newStatus, updBatch.Status)

}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	// настраиваем подключение к tracker.db
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	// задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	// add
	for i := range parcels {
		// добавляем новую посылку в БД, убеждаемся в отсутствии ошибки и наличии идентификатора
		id, err := store.Add(parcels[i])
		require.NoError(t, err)
		require.NotZero(t, id)
		// обновляем идентификатор добавленной у посылки
		parcels[i].Number = id
		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[id] = parcels[i]
	}

	// get by client
	// получаем список посылок по идентификатору клиента, сохранённого в переменной client
	// убеждаемся в отсутствии ошибки
	batch, err := store.GetByClient(client)
	require.NoError(t, err)
	// убеждаемся, что количество полученных посылок совпадает с количеством добавленных
	assert.Len(t, batch, len(parcels))

	// check
	for _, parcel := range batch {
		// в parcelMap лежат добавленные посылки, ключ - идентификатор посылки, значение - сама посылка
		// убеждаемся, что все посылки из storedParcels есть в parcelMap
		// убеждаемся, что значения полей полученных посылок заполнены верно
		p, ok := parcelMap[parcel.Number]
		assert.True(t, ok)
		require.Equal(t, p, parcel)
	}

}
