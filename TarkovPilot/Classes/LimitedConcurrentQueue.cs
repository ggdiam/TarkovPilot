using System;
using System.Collections.Concurrent;
using System.Collections.Generic;
using System.Diagnostics;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace TarkovPilot
{
    public class LimitedConcurrentQueue<T>
    {
        const int DEFAULT_QUEUE_SIZE = 100;
        private readonly ConcurrentQueue<T> _queue = new ConcurrentQueue<T>();
        private readonly int _maxSize;

        public LimitedConcurrentQueue()
        {
            _maxSize = DEFAULT_QUEUE_SIZE;
        }

        public LimitedConcurrentQueue(int maxSize)
        {
            if (maxSize <= 0) { maxSize = DEFAULT_QUEUE_SIZE; }
            _maxSize = maxSize;
        }

        public void Enqueue(T item)
        {
            _queue.Enqueue(item);
            while (_queue.Count > _maxSize)
            {
                _queue.TryDequeue(out _); // Delete oldest item
            }            
        }

        public T Dequeue()
        {
            _queue.TryDequeue(out var result);
            return result;
        }

        public int Count => _queue.Count;

        public List<T> ToList()
        {
            return _queue.ToList();
        }

        public override string ToString()
        {
            return String.Join("\n", _queue);
        }
    }
}
