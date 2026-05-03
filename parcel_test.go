package main

import (
	"database/sql"
	"errors"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()
	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	number, err := store.Add(parcel)
	assert.NoError(t, err)

	// get
	// получите только что добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что значения всех полей в полученном объекте совпадают со значениями полей в переменной parcel
	gotParcel, err := store.Get(number)
	assert.Equal(t, parcel.Client, gotParcel.Client)
	assert.Equal(t, parcel.Status, gotParcel.Status)
	assert.Equal(t, parcel.Address, gotParcel.Address)
	assert.Equal(t, parcel.CreatedAt, gotParcel.CreatedAt)

	// delete
	// удалите добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что посылку больше нельзя получить из БД
	err = store.Delete(number)
	assert.NoError(t, err)

	_, err = store.Get(number)
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()
	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	number, err := store.Add(parcel)
	assert.NoError(t, err)

	// set address
	// обновите адрес, убедитесь в отсутствии ошибки
	newAddress := "new test address"
	err = store.SetAddress(number, newAddress)
	assert.NoError(t, err)

	// check
	// получите добавленную посылку и убедитесь, что адрес обновился
	gotParcel, err := store.Get(number)
	assert.NoError(t, err)
	assert.Equal(t, newAddress, gotParcel.Address)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()
	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	number, err := store.Add(parcel)
	assert.NoError(t, err)

	// set status
	// обновите статус, убедитесь в отсутствии ошибки
	gotOldParcel, err := store.Get(number)
	assert.NoError(t, err)

	var newStatus string
	switch gotOldParcel.Status {
	case ParcelStatusRegistered:
		newStatus = ParcelStatusSent
	case ParcelStatusSent:
		newStatus = ParcelStatusDelivered
	case ParcelStatusDelivered:
		newStatus = ParcelStatusSent
	default:
		require.Error(t, errors.New("unknown status"))
	}

	err = store.SetStatus(number, newStatus)
	assert.NoError(t, err)

	// check
	// получите добавленную посылку и убедитесь, что статус обновился
	gotNewParcel, err := store.Get(number)
	assert.Equal(t, newStatus, gotNewParcel.Status)
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
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
	for i := 0; i < len(parcels); i++ {
		number, err := store.Add(parcels[i])
		assert.NoError(t, err)

		// обновляем идентификатор добавленной у посылки
		parcels[i].Number = int(number)

		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[int(number)] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetByClient(client)
	assert.NoError(t, err)
	require.Equal(t, len(parcels), len(storedParcels))
	// убедитесь в отсутствии ошибки
	// убедитесь, что количество полученных посылок совпадает с количеством добавленных

	// check
	for _, parcel := range storedParcels {
		// в parcelMap лежат добавленные посылки, ключ - идентификатор посылки, значение - сама посылка
		// убедитесь, что все посылки из storedParcels есть в parcelMap
		// убедитесь, что значения полей полученных посылок заполнены верно
		if _, ok := parcelMap[parcel.Number]; !ok {
			t.Errorf("parcel number %d is not found", parcel.Number)
		}
	}
}
